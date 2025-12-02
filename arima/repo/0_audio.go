package repo

import (
	"time"
)

// Audio 音频表（用于分发的音频文件信息）
type Audio struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TrackID      int64      `gorm:"column:track_id;not null" json:"track_id"`
	Size         int64      `gorm:"column:size;not null" json:"size"`
	FileKey      string     `gorm:"column:file_key;type:varchar(255);not null" json:"file_key"`
	Hash         string     `gorm:"column:hash;type:varchar(64);not null" json:"hash"`
	Duration     int        `gorm:"column:duration;not null" json:"duration"`
	Bitrate      int        `gorm:"column:bitrate;not null" json:"bitrate"`
	ChannelCount int        `gorm:"column:channel_count;not null" json:"channel_count"`
	SamplingRate int        `gorm:"column:sampling_rate;not null" json:"sampling_rate"`
	BitDepth     int        `gorm:"column:bit_depth;not null" json:"bit_depth"`
	Format       string     `gorm:"column:format;type:varchar(255);not null" json:"format"`
	Encoder      string     `gorm:"column:encoder;type:varchar(255);not null" json:"encoder"`
	HasLyric     bool       `gorm:"column:has_lyric;not null" json:"has_lyric"`
	HasCover     bool       `gorm:"column:has_cover;not null" json:"has_cover"`
	QualityLabel *string    `gorm:"column:quality_label;type:varchar(255)" json:"quality_label"`
	IsDirty      bool       `gorm:"column:is_dirty;not null;default:false" json:"is_dirty"`
	Source       *string    `gorm:"column:source;type:varchar(255)" json:"source"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamptz" json:"updated_at"`
}

// TableName 指定表名
func (Audio) TableName() string {
	return "audio"
}

// GetAudioList 获取音频列表（分页）
func GetAudioList(page, pageSize int) ([]Audio, int64, error) {
	var audios []Audio
	var total int64

	if err := DB.Model(&Audio{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Order("id").Offset(offset).Limit(pageSize).Find(&audios).Error; err != nil {
		return nil, 0, err
	}

	return audios, total, nil
}

// GetAudioByID 根据ID获取音频
func GetAudioByID(id int64) (Audio, error) {
	var audio Audio
	result := DB.First(&audio, id)
	return audio, result.Error
}

