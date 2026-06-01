package schedule

import (
	"context"
	"fmt"
	"time"
)

type ResourceKey string

const SharedResourceKey = "batch/schedule/shared"

type SharedResource map[string]any

func newSharedResource() SharedResource {
	return make(map[string]any)
}

type Executor[T any] struct {
	job *Job[T]

	repository      Repository
	eventDispatcher *JobEventDispatcher

	clock func() time.Time
}

func NewExecutor[T any](job *Job[T], repository Repository) *Executor[T] {
	return &Executor[T]{
		job:             job,
		repository:      repository,
		eventDispatcher: NewJobEventDispatcher(),
		clock:           time.Now,
	}
}

func (exec *Executor[T]) setClock(clock func() time.Time) {
	exec.clock = clock
}

func (exec *Executor[T]) AddListener(listener JobEventListener) {
	exec.eventDispatcher.Add(listener)
}

func (exec *Executor[T]) Run(ctx context.Context, params JobParameter) error {
	if err := ctx.Err(); err != nil {
		exec.eventDispatcher.On(ctx, NewJobEventCancelled(exec.job.name, exec.clock()))
		return err
	}

	ctx = context.WithValue(ctx, SharedResourceKey, newSharedResource())

	inst, err := exec.repository.Get(ctx, exec.job.name)
	if err != nil {
		exec.eventDispatcher.On(ctx, NewJobEventError(exec.job.name, exec.clock(), err))
		return err
	}

	if inst.State != JobStatusIdle {
		err = fmt.Errorf("job %s is not idle", exec.job.name)
		exec.eventDispatcher.On(ctx, NewJobEventError(exec.job.name, exec.clock(), err))
		return err
	}

	if err = exec.repository.Update(ctx, NewInstance(exec.job.name, JobStatusRunning)); err != nil {
		exec.eventDispatcher.On(ctx, NewJobEventError(exec.job.name, exec.clock(), err))
		return err
	}

	defer func() {
		cleanupCtx := context.WithoutCancel(ctx)
		if updateErr := exec.repository.Update(cleanupCtx, NewInstance(exec.job.name, JobStatusIdle)); updateErr != nil {
			exec.eventDispatcher.On(cleanupCtx, NewJobEventError(exec.job.name, exec.clock(), updateErr))
		}
	}()

	exec.eventDispatcher.On(ctx, NewJobEventStart(exec.job.name, exec.clock()))

	if err = exec.job.Run(ctx, params); err != nil {
		if ctx.Err() != nil {
			exec.eventDispatcher.On(ctx, NewJobEventCancelled(exec.job.name, exec.clock()))
			return ctx.Err()
		}

		exec.eventDispatcher.On(ctx, NewJobEventError(exec.job.name, exec.clock(), err))
		return err
	}

	exec.eventDispatcher.On(ctx, NewJobEventCompleted(exec.job.name, exec.clock()))
	return nil
}

func GetSharedResource(ctx context.Context) SharedResource {
	shared := ctx.Value(SharedResourceKey)
	if shared == nil {
		return nil
	}
	if resource, ok := shared.(SharedResource); ok {
		return resource
	}
	return nil
}
