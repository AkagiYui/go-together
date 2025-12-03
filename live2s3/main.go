package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "config.toml", "配置文件路径")
	flag.Parse()

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	log.Println("视频录制上传系统启动中...")

	// 加载配置
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	log.Printf("配置加载成功: %d个摄像头, S3上传=%v, 代理=%v",
		len(config.Cameras), config.S3.Enabled, config.Proxy.Enabled)

	// 确保存储目录存在
	if err := os.MkdirAll(config.Recording.StoragePath, 0755); err != nil {
		log.Fatalf("创建存储目录失败: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建上传器
	uploader, err := NewUploader(config.S3, config.Proxy, config.Recording.StoragePath)
	if err != nil {
		log.Fatalf("创建上传器失败: %v", err)
	}

	// 启动上传器(3个worker)
	uploader.Start(ctx, 3)

	// 创建存储管理器
	storageManager := NewStorageManager(
		config.Recording.StoragePath,
		config.Recording.MaxDiskUsageGB,
		uploader,
	)
	storageManager.Start(ctx)

	// 启动录制器
	var wg sync.WaitGroup
	var recorders []*Recorder

	for _, cam := range config.Cameras {
		if !cam.Enabled {
			log.Printf("摄像头 %s 已禁用，跳过", cam.Name)
			continue
		}

		recorder := NewRecorder(cam, config.Recording.StoragePath, config.Recording.SegmentDuration)

		// 设置回调：录像完成后加入上传队列
		recorder.SetSegmentCallback(func(filePath string) {
			uploader.QueueUpload(filePath)
		})

		if err := recorder.Start(ctx); err != nil {
			log.Printf("启动摄像头 %s 录制失败: %v", cam.Name, err)
			continue
		}

		recorders = append(recorders, recorder)
		wg.Add(1)
		log.Printf("摄像头 %s 录制已启动", cam.Name)
	}

	if len(recorders) == 0 {
		log.Fatal("没有可用的摄像头，退出")
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("系统运行中，按 Ctrl+C 停止...")

	<-sigChan
	log.Println("收到停止信号，正在停止...")

	// 先停止所有录制器（会优雅终止FFmpeg），然后再取消上下文
	// 注意：必须先 Stop() 再 cancel()，否则 context 取消会导致录制循环提前退出
	var stopWg sync.WaitGroup
	for _, recorder := range recorders {
		stopWg.Add(1)
		go func(r *Recorder) {
			defer stopWg.Done()
			r.Stop()
		}(recorder)
	}
	stopWg.Wait()

	// 最后取消上下文
	cancel()

	log.Println("系统已停止")
}
