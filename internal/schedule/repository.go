package schedule

import "context"

type Repository interface {
	Get(ctx context.Context, name string) (Instance, error)
	Update(context.Context, Instance) error
}

type JobRepository interface {
	Get(ctx context.Context, name string) (Instance, error)
	GetForUpdate(ctx context.Context, name string) (Instance, error)
	Update(ctx context.Context, instance Instance) error
	Transaction(ctx context.Context, fn func(JobRepository) error) error
}
