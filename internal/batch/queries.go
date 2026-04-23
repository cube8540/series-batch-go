package batch

import (
	"context"
	"gorm.io/gorm"
	"series-batch-go/internal/batch/repository"
	"series-batch-go/internal/pkg/llm"
	"time"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindBatchByStatus(ctx context.Context, limit int, status ...llm.JobStatus) []*Batch {
	var ss []string
	for _, s := range status {
		ss = append(ss, string(s))
	}

	entities := repository.FindBatchByStatus(ctx, r.db, limit, ss...)

	var result []*Batch
	for _, entity := range entities {
		batch := ConvertBatchEntityToDomain(entity)
		for _, t := range entity.Targets {
			batch.Targets = append(batch.Targets, ConvertTargetEntityToDomain(&t))
		}
		result = append(result, batch)
	}

	return result
}

func (r *Repository) SaveBatch(ctx context.Context, batches []*Batch) error {
	var entities []*repository.BatchEntity
	for _, b := range batches {
		entity := &repository.BatchEntity{
			ExternalID:   b.ExternalID,
			Type:         string(b.Type),
			Status:       string(b.State),
			RegisteredAt: time.Now(),
		}
		for _, t := range b.Targets {
			entity.Targets = append(entity.Targets, repository.TargetEntity{
				BookID:    t.BookID,
				RequestID: t.RequestID,
			})
		}
		entities = append(entities, entity)
	}
	return repository.SaveBatch(ctx, r.db, entities)
}

func (r *Repository) UpdateBatchStatus(ctx context.Context, ID uint, status llm.JobStatus) error {
	return repository.UpdateBatchStatus(ctx, r.db, ID, string(status))
}

func ConvertBatchEntityToDomain(entity *repository.BatchEntity) *Batch {
	return &Batch{
		ID:           entity.ID,
		ExternalID:   entity.ExternalID,
		Type:         NewType(entity.Type),
		State:        llm.NewJobStatus(entity.Status),
		RegisteredAt: entity.RegisteredAt,
		ModifiedAt:   entity.ModifiedAt,
	}
}

func ConvertTargetEntityToDomain(entity *repository.TargetEntity) Target {
	return Target{
		BookID:    entity.BookID,
		RequestID: entity.RequestID,
	}
}
