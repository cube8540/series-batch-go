package batch

import (
	"time"
)

type jobEntity struct {
	ID        uint `gorm:"primaryKey;autoIncrement:true"`
	Name      string
	State     string
	UpdatedAt time.Time
}

func newJobEntity(instance Job) jobEntity {
	return jobEntity{
		Name:      instance.Name,
		State:     instance.State,
		UpdatedAt: time.Now(),
	}
}

func (e jobEntity) domain() Job {
	return Job{
		ID:        e.ID,
		Name:      e.Name,
		State:     e.State,
		UpdatedAt: e.UpdatedAt,
	}
}

func (e jobEntity) TableName() string {
	return "books.schedule"
}

type batchEntity struct {
	ID         uint `gorm:"primaryKey;autoIncrement:true"`
	ExternalID string
	Type       string
	Status     string

	Targets []targetEntity `gorm:"foreignKey:BatchID;"`

	RegisteredAt time.Time
	ModifiedAt   *time.Time
}

func newBatchEntity(batch *Batch) *batchEntity {
	entity := &batchEntity{
		ExternalID:   batch.ExternalID,
		Type:         string(batch.Type),
		Status:       string(batch.Status),
		RegisteredAt: batch.RegisteredAt,
	}

	if batch.ID != 0 {
		entity.ID = batch.ID
	}

	if batch.Targets != nil {
		for _, target := range batch.Targets {
			entity.Targets = append(entity.Targets, targetEntity{
				BookID:    target.BookID,
				RequestID: target.RequestID,
			})
		}
	}

	return entity
}

func (e *batchEntity) domain() *Batch {
	batch := &Batch{
		ID:           e.ID,
		ExternalID:   e.ExternalID,
		Type:         NewType(e.Type),
		Status:       NewStatus(e.Status),
		RegisteredAt: e.RegisteredAt,
		ModifiedAt:   e.ModifiedAt,
	}

	if e.Targets != nil {
		for _, entity := range e.Targets {
			batch.Targets = append(batch.Targets, entity.domain())
		}
	}

	return batch
}

func (e *batchEntity) TableName() string {
	return "books.series_batch"
}

type targetEntity struct {
	BatchID   uint `gorm:"primaryKey"`
	BookID    uint `gorm:"primaryKey"`
	RequestID string
}

func (e *targetEntity) domain() Target {
	return Target{
		BookID:    e.BookID,
		RequestID: e.RequestID,
	}
}

func (e *targetEntity) TableName() string {
	return "books.series_batch_target"
}
