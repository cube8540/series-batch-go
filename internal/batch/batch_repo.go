package batch

import "context"

type Repository interface {
	Get(ctx context.Context, ID ...uint) []*Batch
	GetByStatus(ctx context.Context, limit int, status ...Status) []*Batch
	Save(ctx context.Context, batches []*Batch) ([]*Batch, error)
	Update(ctx context.Context, batch *Batch) error
	Remove(ctx context.Context, batches []*Batch) error
}

type JobRepository interface {
	Get(ctx context.Context, name string) (Job, error)
	GetForUpdate(ctx context.Context, name string) (Job, error)
	Update(ctx context.Context, job Job) error
	Transaction(ctx context.Context, fn func(JobRepository) error) error
}
