package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/akagiyui/go-together/common/cryptor/encrypt"

	"github.com/akagiyui/go-together/arima/config"
	"github.com/akagiyui/go-together/arima/pkg/ffmpeg"
	"github.com/akagiyui/go-together/arima/pkg/s3"
	"github.com/akagiyui/go-together/arima/repo"
)

// UploadOriginAudioRequest 上传原始音频请求
type UploadOriginAudioRequest struct {
	Files  []*multipart.FileHeader `form:"files"`
	Source *string                 `form:"source"`
}

// Do 处理上传原始音频请求
func (r UploadOriginAudioRequest) Do() (any, error) {
	if len(r.Files) == 0 {
		return nil, fmt.Errorf("no files uploaded")
	}

	var results []repo.OriginAudio
	cfg := config.GlobalConfig
	ff := ffmpeg.NewFFmpeg(cfg.FFmpegExecutable, cfg.FFprobeExecutable)

	for _, fileHeader := range r.Files {
		if fileHeader.Size == 0 {
			continue
		}

		if fileHeader.Size > 100*1024*1024 { // 100MB
			return nil, fmt.Errorf("file size exceeds 100MB limit")
		}

		// 读取文件内容
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return nil, err
		}

		// 计算哈希
		hash := encrypt.Sha256(fileBytes)

		// 检查是否已存在
		count, err := repo.CountOriginAudioByHash(hash)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("file already exists")
		}

		// 保存临时文件用于 FFprobe 分析
		tmpFile, err := os.CreateTemp("", "audio_*"+filepath.Ext(fileHeader.Filename))
		if err != nil {
			return nil, err
		}
		tmpPath := tmpFile.Name()
		_, err = tmpFile.Write(fileBytes)
		tmpFile.Close()
		if err != nil {
			os.Remove(tmpPath)
			return nil, err
		}

		// 分析音频元数据
		probeResult, err := ff.Probe(context.Background(), tmpPath)
		os.Remove(tmpPath)
		if err != nil {
			return nil, err
		}

		// 提取音频流信息
		metadata, err := extractAudioMetadata(probeResult, fileHeader.Filename)
		if err != nil {
			return nil, err
		}

		// 上传到 S3
		fileKey := fmt.Sprintf("origin_audio/%s", hash)
		contentType := fileHeader.Header.Get("Content-Type")
		if err := s3.S3Client.PutObject(context.Background(), fileKey, fileBytes, contentType); err != nil {
			return nil, err
		}

		// 保存数据库记录
		originAudio := repo.OriginAudio{
			FileKey:      fileKey,
			IsRejected:   false,
			FileName:     fileHeader.Filename,
			Size:         fileHeader.Size,
			Duration:     metadata.Duration,
			Hash:         hash,
			Format:       metadata.Format,
			Title:        metadata.Title,
			Artist:       metadata.Artist,
			Album:        metadata.Album,
			Lyrics:       metadata.Lyrics,
			BitRate:      metadata.BitRate,
			ChannelCount: metadata.ChannelCount,
			SampleRate:   metadata.SampleRate,
			Encoding:     metadata.Encoding,
			Encoder:      metadata.Encoder,
			HasCover:     metadata.HasCover,
			Source:       r.Source,
		}

		created, err := repo.CreateOriginAudio(originAudio)
		if err != nil {
			return nil, err
		}
		results = append(results, created)
	}

	return results, nil
}

// Metadata 音频元数据
type Metadata struct {
	Duration     float64
	Format       string
	Title        *string
	Artist       *string
	Album        *string
	Lyrics       json.RawMessage
	BitRate      int
	ChannelCount int
	SampleRate   int
	Encoding     *string
	Encoder      *string
	HasCover     bool
}

func extractAudioMetadata(result *ffmpeg.ProbeResult, _ string) (*Metadata, error) {
	// 查找音频流
	var audioStream *ffmpeg.StreamInfo
	for i := range result.Streams {
		if result.Streams[i].CodecType == "audio" {
			audioStream = &result.Streams[i]
			break
		}
	}
	if audioStream == nil {
		return nil, fmt.Errorf("no audio stream found")
	}

	// 解析时长
	duration, _ := strconv.ParseFloat(result.Format.Duration, 64)

	// 解析比特率
	bitRate, _ := strconv.Atoi(audioStream.BitRate)
	if bitRate == 0 {
		bitRate, _ = strconv.Atoi(result.Format.BitRate)
	}

	// 解析采样率
	sampleRate, _ := strconv.Atoi(audioStream.SampleRate)

	// 检查是否有封面
	hasCover := false
	for _, stream := range result.Streams {
		if stream.CodecType == "video" {
			if comment, ok := stream.Tags["comment"]; ok && strings.HasPrefix(comment, "Cover") {
				hasCover = true
				break
			}
		}
	}

	// 提取标签
	var title, artist, album *string
	if t, ok := result.Format.Tags["title"]; ok {
		title = &t
	}
	if a, ok := result.Format.Tags["artist"]; ok {
		artist = &a
	}
	if al, ok := result.Format.Tags["album"]; ok {
		album = &al
	}

	// 提取歌词
	lyrics := json.RawMessage("{}")
	lyricsMap := make(map[string]string)
	for k, v := range result.Format.Tags {
		if strings.HasPrefix(k, "lyric") {
			lyricsMap[k] = v
		}
	}
	if len(lyricsMap) > 0 {
		lyricsBytes, _ := json.Marshal(lyricsMap)
		lyrics = lyricsBytes
	}

	encoding := &audioStream.CodecName
	var encoder *string
	if enc, ok := audioStream.Tags["encoder"]; ok {
		encoder = &enc
	}

	return &Metadata{
		Duration:     duration,
		Format:       result.Format.FormatName,
		Title:        title,
		Artist:       artist,
		Album:        album,
		Lyrics:       lyrics,
		BitRate:      bitRate,
		ChannelCount: audioStream.Channels,
		SampleRate:   sampleRate,
		Encoding:     encoding,
		Encoder:      encoder,
		HasCover:     hasCover,
	}, nil
}
