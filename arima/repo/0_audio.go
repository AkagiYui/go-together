package repo

import (
	"time"
)

// Audio 音频表（用于分发的音频文件信息）
type Audio struct {
	ID           int64      `json:"id"           gorm:"column:id;primaryKey;autoIncrement"`
	TrackID      int64      `json:"trackId"      gorm:"column:track_id;not null"`
	Size         int64      `json:"size"         gorm:"column:size;not null"`
	FileKey      string     `json:"fileKey"      gorm:"column:file_key;type:varchar(255);not null"`
	Hash         string     `json:"hash"         gorm:"column:hash;type:varchar(64);not null"`
	Duration     int        `json:"duration"     gorm:"column:duration;not null"`
	Bitrate      int        `json:"bitrate"      gorm:"column:bitrate;not null"`
	ChannelCount int        `json:"channelCount" gorm:"column:channel_count;not null"`
	SamplingRate int        `json:"samplingRate" gorm:"column:sampling_rate;not null"`
	BitDepth     int        `json:"bitDepth"     gorm:"column:bit_depth;not null"`
	Format       string     `json:"format"       gorm:"column:format;type:varchar(255);not null"`
	Encoder      string     `json:"encoder"      gorm:"column:encoder;type:varchar(255);not null"`
	HasLyric     bool       `json:"hasLyric"     gorm:"column:has_lyric;not null"`
	HasCover     bool       `json:"hasCover"     gorm:"column:has_cover;not null"`
	QualityLabel *string    `json:"qualityLabel" gorm:"column:quality_label;type:varchar(255)"`
	IsDirty      bool       `json:"isDirty"      gorm:"column:is_dirty;not null;default:false"`
	Source       *string    `json:"source"       gorm:"column:source;type:varchar(255)"`
	CreatedAt    time.Time  `json:"createdAt"    gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp"`
	UpdatedAt    *time.Time `json:"updatedAt"    gorm:"column:updated_at;type:timestamptz"`
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
