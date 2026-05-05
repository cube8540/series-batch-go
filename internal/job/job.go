package job

import (
	"context"
	"errors"
	"slices"
)

type (
	Reader[T any] interface {
		Read(context.Context, map[string]string) []T
	}

	Filter[T any] interface {
		Filter(context.Context, []T) []T
	}

	Processor[I any, O any] interface {
		Process(context.Context, I) O
	}

	Writer[T any] interface {
		Write(context.Context, []T) error
	}
)

type FilterChain[T any] struct {
	n []Filter[T]
}

func NewFilterChain[T any](nodes ...Filter[T]) *FilterChain[T] {
	return &FilterChain[T]{n: nodes}
}

func (c *FilterChain[T]) Filter(ctx context.Context, data []T) []T {
	for _, f := range c.n {
		data = f.Filter(ctx, data)
	}
	return data
}

type ProcessorChain[I any, O any, R any] struct {
	input  Processor[I, O]
	result Processor[O, R]
}

func NewProcessorChain[I any, O any, R any](input Processor[I, O], result Processor[O, R]) *ProcessorChain[I, O, R] {
	return &ProcessorChain[I, O, R]{input: input, result: result}
}

func (c *ProcessorChain[I, O, R]) Process(ctx context.Context, data I) R {
	return c.result.Process(ctx, c.input.Process(ctx, data))
}

type Job[I any, O any] struct {
	reader    Reader[I]
	filter    Filter[I]
	processor Processor[I, O]
	writer    Writer[O]

	chunkSize int
}

func (j *Job[I, O]) Run(ctx context.Context, p map[string]string) error {
	items := j.reader.Read(ctx, p)

	if j.filter != nil {
		items = j.filter.Filter(ctx, items)
	}

	task := func(items []I) error {
		var o []O
		for _, item := range items {
			o = append(o, j.processor.Process(ctx, item))
		}
		return j.writer.Write(ctx, o)
	}

	var errs []error
	for chunk := range slices.Chunk(items, j.chunkSize) {
		errs = append(errs, task(chunk))
	}
	return errors.Join(errs...)
}

type Builder[I any, O any] struct {
	job *Job[I, O]
}

func NewBuilder[I any, O any]() *Builder[I, O] {
	return &Builder[I, O]{job: &Job[I, O]{}}
}

func (b *Builder[I, O]) WithReader(reader Reader[I]) *Builder[I, O] {
	b.job.reader = reader
	return b
}

func (b *Builder[I, O]) WithFilter(filter Filter[I]) *Builder[I, O] {
	b.job.filter = filter
	return b
}

func (b *Builder[I, O]) WithProcessor(processor Processor[I, O]) *Builder[I, O] {
	b.job.processor = processor
	return b
}

func (b *Builder[I, O]) WithWriter(writer Writer[O]) *Builder[I, O] {
	b.job.writer = writer
	return b
}

func (b *Builder[I, O]) WithChunkSize(chunkSize int) *Builder[I, O] {
	b.job.chunkSize = chunkSize
	return b
}

func (b *Builder[I, O]) Build() *Job[I, O] {
	return b.job
}
