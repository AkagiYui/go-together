package repo

import (
	"encoding/json"
	"time"
)

// OriginAudio 原始音频表（上传的原始文件信息）
type OriginAudio struct {
	ID           int64           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FileKey      string          `gorm:"column:file_key;type:varchar(255);not null" json:"fileKey"`
	IsRejected   bool            `gorm:"column:is_rejected;not null;default:false" json:"isRejected"`
	FileName     string          `gorm:"column:file_name;type:varchar(255);not null" json:"fileName"`
	Size         int64           `gorm:"column:size;not null" json:"size"`
	Duration     float64         `gorm:"column:duration;not null" json:"duration"`
	Hash         string          `gorm:"column:hash;type:varchar(64);not null" json:"hash"`
	Format       string          `gorm:"column:format;type:varchar(255);not null" json:"format"`
	Title        *string         `gorm:"column:title;type:varchar(255)" json:"title"`
	Artist       *string         `gorm:"column:artist;type:varchar(255)" json:"artist"`
	Album        *string         `gorm:"column:album;type:varchar(255)" json:"album"`
	Lyrics       json.RawMessage `gorm:"column:lyrics;type:jsonb" json:"lyrics"`
	BitRate      int             `gorm:"column:bit_rate;not null" json:"bitRate"`
	ChannelCount int             `gorm:"column:channel_count;not null" json:"channelCount"`
	SampleRate   int             `gorm:"column:sample_rate;not null" json:"sampleRate"`
	Encoding     *string         `gorm:"column:encoding;type:varchar(255)" json:"encoding"`
	Encoder      *string         `gorm:"column:encoder;type:varchar(255)" json:"encoder"`
	HasCover     bool            `gorm:"column:has_cover;not null" json:"hasCover"`
	Source       *string         `gorm:"column:source;type:varchar(255)" json:"source"`
	CreatedAt    time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp" json:"createdAt"`
	UpdatedAt    *time.Time      `gorm:"column:updated_at;type:timestamptz" json:"updatedAt"`
}

// TableName 指定表名
func (OriginAudio) TableName() string {
	return "origin_audio"
}

// GetOriginAudioList 获取原始音频列表（分页）
func GetOriginAudioList(page, pageSize int) ([]OriginAudio, int64, error) {
	var audios []OriginAudio
	var total int64

	if err := DB.Model(&OriginAudio{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Order("id DESC").Offset(offset).Limit(pageSize).Find(&audios).Error; err != nil {
		return nil, 0, err
	}

	return audios, total, nil
}

// GetOriginAudioByID 根据ID获取原始音频
func GetOriginAudioByID(id int64) (OriginAudio, error) {
	var audio OriginAudio
	result := DB.First(&audio, id)
	return audio, result.Error
}

// GetOriginAudioByHash 根据哈希值获取原始音频
func GetOriginAudioByHash(hash string) (OriginAudio, error) {
	var audio OriginAudio
	result := DB.Where("hash = ?", hash).First(&audio)
	return audio, result.Error
}

// CountOriginAudioByHash 统计指定哈希值的原始音频数量
func CountOriginAudioByHash(hash string) (int64, error) {
	var count int64
	err := DB.Model(&OriginAudio{}).Where("hash = ?", hash).Count(&count).Error
	return count, err
}

// CreateOriginAudio 创建原始音频记录
func CreateOriginAudio(audio OriginAudio) (OriginAudio, error) {
	result := DB.Create(&audio)
	return audio, result.Error
}

