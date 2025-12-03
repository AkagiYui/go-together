package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// 上传相关常量
const (
	maxRetryCount  = 5                // 最大重试次数
	retryBaseDelay = 30 * time.Second // 基础重试延迟
)

// UploadTask 上传任务
type UploadTask struct {
	FilePath   string
	RetryCount int
}

// Uploader S3上传器
type Uploader struct {
	s3Client       *s3.Client
	bucket         string
	enabled        bool
	uploadQueue    chan UploadTask
	uploadedFiles  map[string]bool
	uploadedMutex  sync.RWMutex
	failedFiles    map[string]int // 记录失败次数
	failedMutex    sync.RWMutex
	storagePath    string
	statusFilePath string
	failedFilePath string // 失败文件记录路径
}

// NewUploader 创建新的上传器
func NewUploader(s3Config S3Config, proxyConfig ProxyConfig, storagePath string) (*Uploader, error) {
	uploader := &Uploader{
		bucket:         s3Config.Bucket,
		enabled:        s3Config.Enabled,
		uploadQueue:    make(chan UploadTask, 1000),
		uploadedFiles:  make(map[string]bool),
		failedFiles:    make(map[string]int),
		storagePath:    storagePath,
		statusFilePath: filepath.Join(storagePath, ".uploaded"),
		failedFilePath: filepath.Join(storagePath, ".failed"),
	}

	if !s3Config.Enabled {
		log.Println("S3上传已禁用")
		return uploader, nil
	}

	// 创建HTTP客户端(支持代理)
	httpClient := &http.Client{
		Timeout: 30 * time.Minute, // 上传可能需要较长时间
	}

	if proxyConfig.Enabled {
		proxyURL := proxyConfig.GetProxyURL()
		if proxyURL != "" {
			parsedURL, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("解析代理地址失败: %w", err)
			}
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(parsedURL),
			}
			log.Printf("S3上传将使用代理: %s", proxyConfig.Address)
		}
	}

	// 创建S3客户端
	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(s3Config.Endpoint),
		Region:       s3Config.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(s3Config.AccessKey, s3Config.SecretKey, ""),
		HTTPClient:   httpClient,
	})

	uploader.s3Client = s3Client

	// 加载已上传文件状态
	uploader.loadUploadedStatus()

	return uploader, nil
}

// loadUploadedStatus 加载已上传文件状态
func (u *Uploader) loadUploadedStatus() {
	file, err := os.Open(u.statusFilePath)
	if err != nil {
		return
	}
	defer file.Close()

	u.uploadedMutex.Lock()
	defer u.uploadedMutex.Unlock()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			u.uploadedFiles[line] = true
		}
	}
}

// saveUploadedStatus 保存已上传文件状态
func (u *Uploader) saveUploadedStatus() error {
	u.uploadedMutex.RLock()
	defer u.uploadedMutex.RUnlock()

	var content string
	for file := range u.uploadedFiles {
		content += file + "\n"
	}

	return os.WriteFile(u.statusFilePath, []byte(content), 0644)
}

// IsUploaded 检查文件是否已上传
func (u *Uploader) IsUploaded(filePath string) bool {
	u.uploadedMutex.RLock()
	defer u.uploadedMutex.RUnlock()
	return u.uploadedFiles[filePath]
}

// MarkUploaded 标记文件为已上传
func (u *Uploader) MarkUploaded(filePath string) {
	u.uploadedMutex.Lock()
	u.uploadedFiles[filePath] = true
	u.uploadedMutex.Unlock()
	u.saveUploadedStatus()
}

// QueueUpload 将文件加入上传队列
func (u *Uploader) QueueUpload(filePath string) {
	u.queueUploadWithRetry(filePath, 0)
}

// queueUploadWithRetry 带重试次数的上传队列
func (u *Uploader) queueUploadWithRetry(filePath string, retryCount int) {
	if !u.enabled {
		return
	}
	task := UploadTask{
		FilePath:   filePath,
		RetryCount: retryCount,
	}
	select {
	case u.uploadQueue <- task:
		if retryCount == 0 {
			log.Printf("文件已加入上传队列: %s", filePath)
		} else {
			log.Printf("文件重新加入上传队列 (重试 %d/%d): %s", retryCount, maxRetryCount, filePath)
		}
	default:
		log.Printf("上传队列已满，跳过文件: %s", filePath)
	}
}

// Start 启动上传worker
func (u *Uploader) Start(ctx context.Context, workerCount int) {
	if !u.enabled {
		return
	}

	// 首先扫描并上传未上传的文件
	go u.scanAndUploadPending(ctx)

	// 启动上传worker
	for i := range workerCount {
		go u.uploadWorker(ctx, i)
	}
}

// scanAndUploadPending 扫描并上传未上传的文件
func (u *Uploader) scanAndUploadPending(_ context.Context) {
	// 等待一段时间让系统稳定
	time.Sleep(5 * time.Second)

	log.Println("开始扫描未上传的文件...")

	err := filepath.Walk(u.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".mp4" {
			return nil
		}

		// 检查文件是否已上传
		if !u.IsUploaded(path) {
			// 确保文件已完成写入
			if time.Since(info.ModTime()) > 30*time.Second {
				u.QueueUpload(path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("扫描文件失败: %v", err)
	}
}

// uploadWorker 上传工作协程
func (u *Uploader) uploadWorker(ctx context.Context, id int) {
	log.Printf("上传Worker #%d 已启动", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("上传Worker #%d 已停止", id)
			return
		case task := <-u.uploadQueue:
			if u.IsUploaded(task.FilePath) {
				continue
			}

			if err := u.uploadFile(ctx, task.FilePath); err != nil {
				log.Printf("上传失败 [%s]: %v", task.FilePath, err)

				// 检查是否超过最大重试次数
				if task.RetryCount >= maxRetryCount {
					log.Printf("上传失败次数已达上限 (%d次)，放弃上传: %s", maxRetryCount, task.FilePath)
					u.markFailed(task.FilePath)
				} else {
					// 重新加入队列，延迟重试（指数退避）
					nextRetry := task.RetryCount + 1
					delay := retryBaseDelay * time.Duration(1<<uint(task.RetryCount)) // 30s, 60s, 120s, 240s, 480s
					go func(path string, retry int, d time.Duration) {
						log.Printf("将在 %v 后重试上传: %s", d, path)
						time.Sleep(d)
						u.queueUploadWithRetry(path, retry)
					}(task.FilePath, nextRetry, delay)
				}
			} else {
				u.MarkUploaded(task.FilePath)
				log.Printf("上传成功: %s", task.FilePath)
			}
		}
	}
}

// markFailed 标记文件上传失败
func (u *Uploader) markFailed(filePath string) {
	u.failedMutex.Lock()
	u.failedFiles[filePath] = maxRetryCount
	u.failedMutex.Unlock()

	// 记录到失败文件列表
	u.saveFailedFile(filePath)
}

// saveFailedFile 保存失败文件到列表
func (u *Uploader) saveFailedFile(filePath string) {
	f, err := os.OpenFile(u.failedFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("无法记录失败文件: %v", err)
		return
	}
	defer f.Close()
	f.WriteString(filePath + "\n")
}

// IsFailed 检查文件是否已标记为失败
func (u *Uploader) IsFailed(filePath string) bool {
	u.failedMutex.RLock()
	defer u.failedMutex.RUnlock()
	_, exists := u.failedFiles[filePath]
	return exists
}

// GetFailedFiles 获取所有失败的文件列表
func (u *Uploader) GetFailedFiles() []string {
	u.failedMutex.RLock()
	defer u.failedMutex.RUnlock()
	files := make([]string, 0, len(u.failedFiles))
	for f := range u.failedFiles {
		files = append(files, f)
	}
	return files
}

// uploadFile 上传单个文件到S3
func (u *Uploader) uploadFile(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 计算S3 key: 使用相对路径
	relPath, err := filepath.Rel(u.storagePath, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	log.Printf("正在上传: %s (%.2f MB)", relPath, float64(fileInfo.Size())/1024/1024)

	_, err = u.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(u.bucket),
		Key:           aws.String(relPath),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("video/mp4"),
	})

	if err != nil {
		return fmt.Errorf("上传到S3失败: %w", err)
	}

	return nil
}
