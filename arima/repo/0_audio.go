package repo

import (
	"time"
)

// Audio 音频表（用于分发的音频文件信息）
type Audio struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TrackID      int64      `gorm:"column:track_id;not null" json:"trackId"`
	Size         int64      `gorm:"column:size;not null" json:"size"`
	FileKey      string     `gorm:"column:file_key;type:varchar(255);not null" json:"fileKey"`
	Hash         string     `gorm:"column:hash;type:varchar(64);not null" json:"hash"`
	Duration     int        `gorm:"column:duration;not null" json:"duration"`
	Bitrate      int        `gorm:"column:bitrate;not null" json:"bitrate"`
	ChannelCount int        `gorm:"column:channel_count;not null" json:"channelCount"`
	SamplingRate int        `gorm:"column:sampling_rate;not null" json:"samplingRate"`
	BitDepth     int        `gorm:"column:bit_depth;not null" json:"bitDepth"`
	Format       string     `gorm:"column:format;type:varchar(255);not null" json:"format"`
	Encoder      string     `gorm:"column:encoder;type:varchar(255);not null" json:"encoder"`
	HasLyric     bool       `gorm:"column:has_lyric;not null" json:"hasLyric"`
	HasCover     bool       `gorm:"column:has_cover;not null" json:"hasCover"`
	QualityLabel *string    `gorm:"column:quality_label;type:varchar(255)" json:"qualityLabel"`
	IsDirty      bool       `gorm:"column:is_dirty;not null;default:false" json:"isDirty"`
	Source       *string    `gorm:"column:source;type:varchar(255)" json:"source"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp" json:"createdAt"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamptz" json:"updatedAt"`
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
