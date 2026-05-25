package batch

import (
	"context"
	"series-batch-go/internal/config/log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (repo GormRepository) GetByStatus(ctx context.Context, limit int, status ...Status) []*Batch {
	var statusRaw []string
	for _, s := range status {
		statusRaw = append(statusRaw, string(s))
	}

	entities, err := gorm.G[*batchEntity](repo.db.Unscoped()).Preload("Targets", nil).Where("status IN ?", statusRaw).Order("id desc").Limit(limit).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding batch(%v): %v", status, err)
	}

	var result []*Batch
	for _, entity := range entities {
		result = append(result, entity.domain())
	}
	return result
}

func (repo GormRepository) Save(ctx context.Context, batches []*Batch) ([]*Batch, error) {
	var entities []*batchEntity
	for _, batch := range batches {
		entities = append(entities, newBatchEntity(batch))
	}

	resp := repo.db.WithContext(ctx).Create(&entities)
	if resp.Error != nil {
		return nil, resp.Error
	}

	var result []*Batch
	for _, entity := range entities {
		result = append(result, entity.domain())
	}
	return result, nil
}

func (repo GormRepository) Update(ctx context.Context, batch *Batch) error {
	entity, now := newBatchEntity(batch), time.Now()
	entity.ModifiedAt = &now

	_, err := gorm.G[*batchEntity](repo.db).Omit("Targets").Where("id = ?", batch.ID).Updates(ctx, entity)
	return err
}

type GormJobRepository struct {
	db *gorm.DB
}

func NewGormJobRepository(db *gorm.DB) *GormJobRepository {
	return &GormJobRepository{db: db}
}

func (repo GormJobRepository) Get(ctx context.Context, name string) (Job, error) {
	entity, err := gorm.G[jobEntity](repo.db).Where("name = ?", name).First(ctx)
	if err != nil {
		return Job{}, err
	}
	return entity.domain(), nil
}

func (repo GormJobRepository) GetForUpdate(ctx context.Context, name string) (Job, error) {
	db := repo.db.Clauses(clause.Locking{
		Strength: "UPDATE",
		Table:    clause.Table{Name: clause.CurrentTable},
		Options:  "NOWAIT",
	})
	entity, err := gorm.G[jobEntity](db).Where("name = ?", name).First(ctx)
	if err != nil {
		return Job{}, err
	}
	return entity.domain(), nil
}

func (repo GormJobRepository) Update(ctx context.Context, job Job) error {
	entity := newJobEntity(job)
	_, err := gorm.G[jobEntity](repo.db).Where("name = ?", job.Name).Updates(ctx, entity)
	return err
}

func (repo GormJobRepository) Transaction(ctx context.Context, fn func(JobRepository) error) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewGormJobRepository(tx))
	})
}
