package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// StorageManager 存储管理器
type StorageManager struct {
	storagePath    string
	maxDiskUsageGB int
	uploader       *Uploader
}

// NewStorageManager 创建存储管理器
func NewStorageManager(storagePath string, maxDiskUsageGB int, uploader *Uploader) *StorageManager {
	return &StorageManager{
		storagePath:    storagePath,
		maxDiskUsageGB: maxDiskUsageGB,
		uploader:       uploader,
	}
}

// Start 启动存储管理
func (s *StorageManager) Start(ctx context.Context) {
	go s.cleanupLoop(ctx)
}

// cleanupLoop 清理循环
func (s *StorageManager) cleanupLoop(ctx context.Context) {
	// 每5分钟检查一次
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 启动时先检查一次
	s.cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup 执行清理
func (s *StorageManager) cleanup() {
	currentUsage, err := s.getDiskUsage()
	if err != nil {
		log.Printf("获取磁盘使用量失败: %v", err)
		return
	}

	maxUsageBytes := int64(s.maxDiskUsageGB) * 1024 * 1024 * 1024
	if currentUsage < maxUsageBytes {
		return // 未超过阈值
	}

	log.Printf("磁盘使用量 %.2f GB 超过阈值 %d GB，开始清理...",
		float64(currentUsage)/1024/1024/1024, s.maxDiskUsageGB)

	// 获取所有文件并按时间排序
	files, err := s.getAllFiles()
	if err != nil {
		log.Printf("获取文件列表失败: %v", err)
		return
	}

	// 按修改时间排序(最旧的在前)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	// 删除已上传的最旧文件直到磁盘使用量低于阈值
	for _, file := range files {
		if currentUsage < maxUsageBytes {
			break
		}

		// 只删除已上传的文件
		if s.uploader != nil && !s.uploader.IsUploaded(file.Path) {
			log.Printf("跳过未上传文件: %s", file.Path)
			continue
		}

		if err := os.Remove(file.Path); err != nil {
			log.Printf("删除文件失败 [%s]: %v", file.Path, err)
			continue
		}

		currentUsage -= file.Size
		log.Printf("已删除文件: %s (释放 %.2f MB)", file.Path, float64(file.Size)/1024/1024)
	}

	log.Printf("清理完成，当前磁盘使用量: %.2f GB", float64(currentUsage)/1024/1024/1024)
}

// FileInfo 文件信息
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// getDiskUsage 获取存储目录的磁盘使用量
func (s *StorageManager) getDiskUsage() (int64, error) {
	var totalSize int64

	err := filepath.Walk(s.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// getAllFiles 获取所有MP4文件
func (s *StorageManager) getAllFiles() ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(s.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".mp4" {
			return nil
		}

		files = append(files, FileInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})

	return files, err
}
