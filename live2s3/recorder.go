package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Recorder 录制器
type Recorder struct {
	camera          CameraConfig
	storagePath     string
	segmentDuration int
	cmd             *exec.Cmd
	cmdMutex        sync.Mutex // 保护cmd的并发访问
	cancel          context.CancelFunc
	stopChan        chan struct{}         // 用于通知录制循环停止
	stoppedChan     chan struct{}         // 用于确认录制已完全停止
	onSegmentDone   func(filePath string) // 分段完成回调
}

// NewRecorder 创建新的录制器
func NewRecorder(camera CameraConfig, storagePath string, segmentDuration int) *Recorder {
	return &Recorder{
		camera:          camera,
		storagePath:     storagePath,
		segmentDuration: segmentDuration,
		stopChan:        make(chan struct{}),
		stoppedChan:     make(chan struct{}),
	}
}

// SetSegmentCallback 设置分段完成回调
func (r *Recorder) SetSegmentCallback(callback func(filePath string)) {
	r.onSegmentDone = callback
}

// calculateNextAlignedTime 计算下一个对齐的时间点信息
// 返回：到下一个对齐时间点的秒数
func (r *Recorder) calculateSecondsToNextAlignment() int {
	now := time.Now()
	// 计算当天已经过去的秒数
	secondsOfDay := now.Hour()*3600 + now.Minute()*60 + now.Second()
	// 计算到下一个分段点的秒数
	remainder := secondsOfDay % r.segmentDuration
	if remainder == 0 {
		return r.segmentDuration
	}
	return r.segmentDuration - remainder
}

// Start 开始录制
func (r *Recorder) Start(ctx context.Context) error {
	// 确保存储目录存在
	cameraDir := filepath.Join(r.storagePath, r.camera.Name)
	if err := os.MkdirAll(cameraDir, 0755); err != nil {
		return fmt.Errorf("创建存储目录失败: %w", err)
	}

	ctx, r.cancel = context.WithCancel(ctx)

	go r.recordLoop(ctx, cameraDir)
	return nil
}

// Stop 停止录制（优雅退出）
func (r *Recorder) Stop() {
	log.Printf("[%s] 正在停止录制...", r.camera.Name)

	// 通知录制循环停止
	close(r.stopChan)

	// 优雅终止FFmpeg进程
	r.gracefulStopFFmpeg()

	// 取消context
	if r.cancel != nil {
		r.cancel()
	}

	// 等待录制完全停止（最多等待15秒）
	select {
	case <-r.stoppedChan:
		log.Printf("[%s] 录制已完全停止", r.camera.Name)
	case <-time.After(15 * time.Second):
		log.Printf("[%s] 等待录制停止超时", r.camera.Name)
	}
}

// gracefulStopFFmpeg 优雅停止FFmpeg进程
func (r *Recorder) gracefulStopFFmpeg() {
	r.cmdMutex.Lock()
	cmd := r.cmd
	r.cmdMutex.Unlock()

	if cmd == nil || cmd.Process == nil {
		log.Printf("[%s] FFmpeg进程不存在，无需停止", r.camera.Name)
		return
	}

	// 发送SIGINT信号，让FFmpeg优雅退出
	// FFmpeg收到SIGINT后会完成当前文件的写入和封装
	log.Printf("[%s] 发送SIGINT信号给FFmpeg进程...", r.camera.Name)
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		log.Printf("[%s] 发送SIGINT失败: %v，进程可能已退出", r.camera.Name, err)
		return
	}

	// 注意：不需要在这里调用 cmd.Wait()
	// 因为 runFFmpeg() 中的 cmd.Run() 已经在等待进程结束
	// recordLoop 会在 cmd.Run() 返回后自动退出并关闭 stoppedChan
	log.Printf("[%s] 已发送SIGINT信号，等待FFmpeg完成文件写入...", r.camera.Name)
}

// recordLoop 录制循环
func (r *Recorder) recordLoop(ctx context.Context, cameraDir string) {
	defer close(r.stoppedChan) // 退出时通知已停止

	for {
		select {
		case <-r.stopChan:
			log.Printf("[%s] 收到停止信号，退出录制循环", r.camera.Name)
			return
		case <-ctx.Done():
			log.Printf("[%s] Context已取消，退出录制循环", r.camera.Name)
			return
		default:
			if err := r.runFFmpeg(ctx, cameraDir); err != nil {
				// 检查是否是正常停止
				select {
				case <-r.stopChan:
					return
				default:
				}
				log.Printf("[%s] FFmpeg录制错误: %v, 5秒后重试", r.camera.Name, err)
				select {
				case <-r.stopChan:
					return
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
			}
		}
	}
}

// runFFmpeg 运行FFmpeg进程
func (r *Recorder) runFFmpeg(ctx context.Context, cameraDir string) error {
	// 计算到下一个对齐时间点的秒数
	secondsToNextAlignment := r.calculateSecondsToNextAlignment()
	log.Printf("[%s] 距离下一个整%d分钟时间点: %d秒",
		r.camera.Name, r.segmentDuration/60, secondsToNextAlignment)

	// 输出文件名模板: 摄像头名称-年-月-日-时-分.mp4
	outputPattern := filepath.Join(cameraDir, r.camera.Name+"-%Y-%m-%d-%H-%M.mp4")

	// 构建FFmpeg命令
	// 使用segment muxer实现按时钟时间对齐切割
	// 重新编码以确保能在精确时间点切割（copy模式只能在关键帧切割）
	args := []string{
		"-rtsp_transport", "tcp", // 使用TCP传输RTSP
		"-i", r.camera.RtspURL, // 输入RTSP流

		// 视频编码：H.265 (HEVC)
		"-c:v", "libx265",
		"-preset", "fast", // 编码速度预设（ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow）
		"-crf", "23", // 恒定质量因子（0-51，越小质量越高，23是默认值）
		"-tag:v", "hvc1", // 兼容性标签，用于Apple设备播放

		// 音频编码：AAC
		"-c:a", "aac",
		"-b:a", "128k", // 音频比特率

		// 分段设置
		"-f", "segment", // 使用segment muxer
		"-segment_time", fmt.Sprintf("%d", r.segmentDuration), // 分段时长（秒）
		"-segment_atclocktime", "1", // 按实际时钟时间对齐切割
		"-strftime", "1", // 使用strftime格式化文件名
		"-reset_timestamps", "1", // 重置每个片段的时间戳
		"-segment_format", "mp4", // 输出格式为MP4
		"-movflags", "+faststart", // 优化MP4结构，将moov原子移到文件开头

		// 强制在分段边界生成关键帧
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", r.segmentDuration),

		outputPattern,
	}

	// 创建FFmpeg命令（不使用CommandContext，因为我们需要自己控制优雅退出）
	r.cmdMutex.Lock()
	r.cmd = exec.Command("ffmpeg", args...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	// 让FFmpeg在独立的进程组中运行
	// 这样终端的Ctrl+C (SIGINT) 不会直接发送给FFmpeg，我们可以完全控制信号发送
	r.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	r.cmdMutex.Unlock()

	log.Printf("[%s] 启动FFmpeg录制: ffmpeg %v", r.camera.Name, args)

	// 启动文件监控goroutine
	go r.monitorNewFiles(ctx, cameraDir)

	// 运行FFmpeg
	err := r.cmd.Run()

	// 清理cmd引用
	r.cmdMutex.Lock()
	r.cmd = nil
	r.cmdMutex.Unlock()

	// 检查是否是正常停止
	select {
	case <-r.stopChan:
		return nil // 正常停止，不返回错误
	default:
	}

	return err
}

// monitorNewFiles 监控新生成的文件
// 核心逻辑：当检测到新文件出现时，说明前一个文件已经写入完成
func (r *Recorder) monitorNewFiles(ctx context.Context, cameraDir string) {
	// 记录已知文件及其状态
	// true = 已完成（已触发回调）, false = 正在录制中
	knownFiles := make(map[string]bool)

	// 初始化：获取当前所有文件，标记为已完成（避免重复处理历史文件）
	files, _ := filepath.Glob(filepath.Join(cameraDir, "*.mp4"))
	for _, f := range files {
		knownFiles[f] = true // 标记为已完成
	}

	// 当前正在录制的文件路径
	var currentRecordingFile string

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			files, err := filepath.Glob(filepath.Join(cameraDir, "*.mp4"))
			if err != nil {
				continue
			}

			// 按修改时间排序，找出最新的文件（当前正在录制的文件）
			var latestFile string
			var latestModTime time.Time
			for _, f := range files {
				info, err := os.Stat(f)
				if err != nil {
					continue
				}
				if info.ModTime().After(latestModTime) {
					latestModTime = info.ModTime()
					latestFile = f
				}
			}

			// 如果发现了新的"当前录制文件"
			if latestFile != "" && latestFile != currentRecordingFile {
				// 前一个正在录制的文件现在已经完成了
				if currentRecordingFile != "" {
					if completed, exists := knownFiles[currentRecordingFile]; exists && !completed {
						knownFiles[currentRecordingFile] = true // 标记为已完成
						log.Printf("[%s] 录像文件完成: %s", r.camera.Name, currentRecordingFile)
						if r.onSegmentDone != nil {
							r.onSegmentDone(currentRecordingFile)
						}
					}
				}

				// 更新当前录制文件
				currentRecordingFile = latestFile
				if _, exists := knownFiles[latestFile]; !exists {
					knownFiles[latestFile] = false // 标记为正在录制
					log.Printf("[%s] 开始录制新文件: %s", r.camera.Name, filepath.Base(latestFile))
				}
			}
		}
	}
}
