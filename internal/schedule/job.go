package schedule

import (
	"context"
	"errors"
	"slices"
)

// JobDefaultChunkSize 잡 실행시 읽은 데이터를 처리하는 기본 청크 사이즈
const JobDefaultChunkSize = 100

type JobParameter map[string]string

type (
	Reader[T any] interface {
		Read(context.Context, JobParameter) ([]T, error)
	}

	Filter[T any] interface {
		Filter(context.Context, T) bool
	}

	Writer[T any] interface {
		Write(context.Context, []T) error
	}
)

type AcceptAllFilter[T any] struct{}

func NewAcceptAllFilter[T any]() AcceptAllFilter[T] {
	return AcceptAllFilter[T]{}
}

func (f AcceptAllFilter[T]) Filter(_ context.Context, _ T) bool {
	return true
}

type FilterChain[T any] struct {
	filters []Filter[T]
}

func NewFilterChain[T any](filters ...Filter[T]) *FilterChain[T] {
	return &FilterChain[T]{filters: filters}
}

func (c *FilterChain[T]) Filter(ctx context.Context, input T) bool {
	for _, f := range c.filters {
		if !f.Filter(ctx, input) {
			return false
		}
	}
	return true
}

type Job[T any] struct {
	name string

	reader Reader[T]
	filter Filter[T]
	writer Writer[T]

	chunkSize int
}

func (i *Job[T]) Run(ctx context.Context, params JobParameter) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	items, err := i.reader.Read(ctx, params)
	if err != nil {
		return err
	}

	var targets []T
	for _, item := range items {
		if err = ctx.Err(); err != nil {
			return err
		}
		if i.filter.Filter(ctx, item) {
			targets = append(targets, item)
		}
	}

	for chunk := range slices.Chunk(targets, i.chunkSize) {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = i.writer.Write(ctx, chunk); err != nil {
			return err
		}
	}

	return nil
}

type JobBuilder[T any] struct {
	instance *Job[T]
}

func NewJobBuilder[I any](name string) *JobBuilder[I] {
	instance := &Job[I]{
		name:      name,
		chunkSize: JobDefaultChunkSize,
		filter:    NewAcceptAllFilter[I](),
	}
	return &JobBuilder[I]{instance}
}

func (b *JobBuilder[T]) WithReader(reader Reader[T]) *JobBuilder[T] {
	b.instance.reader = reader
	return b
}

func (b *JobBuilder[T]) WithFilter(filter Filter[T]) *JobBuilder[T] {
	b.instance.filter = filter
	return b
}

func (b *JobBuilder[T]) WithWriter(writer Writer[T]) *JobBuilder[T] {
	b.instance.writer = writer
	return b
}

func (b *JobBuilder[T]) WithChunkSize(chunkSize int) *JobBuilder[T] {
	b.instance.chunkSize = chunkSize
	return b
}

func (b *JobBuilder[T]) Build() (*Job[T], error) {
	if b.instance.reader == nil {
		return nil, errors.New("reader is required")
	}
	if b.instance.writer == nil {
		return nil, errors.New("writer is required")
	}
	if b.instance.filter == nil {
		return nil, errors.New("filter is required")
	}
	if b.instance.name == "" {
		return nil, errors.New("name is required")
	}
	if b.instance.chunkSize <= 0 {
		return nil, errors.New("chunkSize must be greater than 0")
	}
	return b.instance, nil
}
