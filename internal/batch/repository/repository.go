package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"series-batch-go/internal/config/log"
	"time"
)

func FindBatchByStatus(ctx context.Context, db *gorm.DB, limit int, s ...string) []*BatchEntity {
	entities, err := gorm.G[*BatchEntity](db.Unscoped()).Preload("Targets", nil).Where("status IN ?", s).Order("id desc").Limit(limit).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding batch(%v): %v", s, err)
	}
	return entities
}

func SaveBatch(ctx context.Context, db *gorm.DB, batch []*BatchEntity) error {
	result := db.WithContext(ctx).Create(batch)
	if result.Error != nil {
		return fmt.Errorf("error occurred when saving batch: %v", result.Error)
	}
	return nil
}

func UpdateBatchStatus(ctx context.Context, db *gorm.DB, ID uint, status string) error {
	t := time.Now()
	_, err := gorm.G[BatchEntity](db).Where("id = ?", ID).Updates(ctx, BatchEntity{Status: status, ModifiedAt: &t})
	return err
}
