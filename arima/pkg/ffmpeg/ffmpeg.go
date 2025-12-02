// Package ffmpeg 提供 FFmpeg 相关功能
package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// FFmpeg FFmpeg 工具封装
type FFmpeg struct {
	FFmpegPath  string
	FFprobePath string
}

// NewFFmpeg 创建 FFmpeg 实例
func NewFFmpeg(ffmpegPath, ffprobePath string) *FFmpeg {
	return &FFmpeg{
		FFmpegPath:  ffmpegPath,
		FFprobePath: ffprobePath,
	}
}

// FFmpegVersion 获取 FFmpeg 版本
func (f *FFmpeg) FFmpegVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, f.FFmpegPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", nil
}

// FFprobeVersion 获取 FFprobe 版本
func (f *FFmpeg) FFprobeVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, f.FFprobePath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", nil
}

// ProbeResult FFprobe 探测结果
type ProbeResult struct {
	Streams []StreamInfo `json:"streams"`
	Format  FormatInfo   `json:"format"`
}

// StreamInfo 流信息
type StreamInfo struct {
	Index          int               `json:"index"`
	CodecName      string            `json:"codec_name"`
	CodecType      string            `json:"codec_type"`
	BitRate        string            `json:"bit_rate,omitempty"`
	Channels       int               `json:"channels,omitempty"`
	SampleRate     string            `json:"sample_rate,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	Width          int               `json:"width,omitempty"`
	Height         int               `json:"height,omitempty"`
	Duration       string            `json:"duration,omitempty"`
	BitsPerSample  int               `json:"bits_per_sample,omitempty"`
	ChannelLayout  string            `json:"channel_layout,omitempty"`
}

// FormatInfo 格式信息
type FormatInfo struct {
	Filename   string            `json:"filename"`
	FormatName string            `json:"format_name"`
	Duration   string            `json:"duration"`
	Size       string            `json:"size"`
	BitRate    string            `json:"bit_rate"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// Probe 使用 FFprobe 探测文件
func (f *FFmpeg) Probe(ctx context.Context, filePath string) (*ProbeResult, error) {
	cmd := exec.CommandContext(ctx, f.FFprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe error: %v, stderr: %s", err, stderr.String())
	}

	var result ProbeResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	return &result, nil
}

