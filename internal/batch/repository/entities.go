package repository

import (
	"time"
)

type BatchEntity struct {
	ID         uint `gorm:"primaryKey;autoIncrement:true"`
	ExternalID string
	Type       string
	Status     string

	Targets []TargetEntity `gorm:"foreignKey:BatchID"`

	RegisteredAt time.Time
	ModifiedAt   *time.Time
}

func (e *BatchEntity) TableName() string {
	return "books.series_batch"
}

type TargetEntity struct {
	BatchID   uint `gorm:"primaryKey"`
	BookID    uint `gorm:"primaryKey"`
	RequestID string
}

func (e *TargetEntity) TableName() string {
	return "books.series_batch_target"
}
