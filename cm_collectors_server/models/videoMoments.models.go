package models

import "time"

// VideoMoment stores positions only. Missing files retain bookmarks; permanent resource deletion removes them.
type VideoMoment struct {
	ID            string    `json:"id" gorm:"primaryKey;type:char(20)"`
	DramaSeriesID string    `json:"dramaSeriesId" gorm:"index;type:char(20)"`
	ResourceID    string    `json:"resourceId" gorm:"index;type:char(20)"`
	Title         string    `json:"title" gorm:"type:varchar(200)"`
	Note          string    `json:"note" gorm:"type:text"`
	Start         float64   `json:"start"`
	End           float64   `json:"end"`
	SourceKey     string    `json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Available     bool      `json:"available" gorm:"-"`
	SourceChanged bool      `json:"sourceChanged" gorm:"-"`
	ResourceTitle string    `json:"resourceTitle" gorm:"-"`
}

type VideoPreviewSetting struct {
	ID    uint  `json:"id" gorm:"primaryKey"`
	MaxMB int64 `json:"maxMB"`
	Width int   `json:"width"`
}

type VideoPreview struct {
	Width int `json:"width"`

	DramaSeriesID string    `json:"dramaSeriesId" gorm:"primaryKey;type:char(20)"`
	SourceKey     string    `json:"sourceKey"`
	Status        string    `json:"status" gorm:"index;type:varchar(20)"`
	Error         string    `json:"error"`
	Interval      int       `json:"interval"`
	Duration      float64   `json:"duration"`
	Count         int       `json:"count"`
	Completed     int       `json:"completed"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
