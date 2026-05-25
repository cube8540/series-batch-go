package schedule

import (
	"context"
	"time"
)

type (
	JobEvent interface {
		JobName() string
		Timestamp() time.Time
	}

	JobEventStart struct {
		name      string
		timestamp time.Time
	}
	JobEventCompleted struct {
		name      string
		timestamp time.Time
	}
	JobEventCancelled struct {
		name      string
		timestamp time.Time
	}
	JobEventError struct {
		name      string
		timestamp time.Time
		err       error
	}
)

func NewJobEventStart(name string, timestamp time.Time) JobEventStart {
	return JobEventStart{name: name, timestamp: timestamp}
}

func (e JobEventStart) JobName() string {
	return e.name
}

func (e JobEventStart) Timestamp() time.Time {
	return e.timestamp
}

func NewJobEventError(name string, timestamp time.Time, err error) JobEventError {
	return JobEventError{name: name, timestamp: timestamp, err: err}
}

func (e JobEventError) JobName() string {
	return e.name
}

func (e JobEventError) Timestamp() time.Time {
	return e.timestamp
}

func (e JobEventError) Error() error {
	return e.err
}

func NewJobEventCancelled(name string, timestamp time.Time) JobEventCancelled {
	return JobEventCancelled{name: name, timestamp: timestamp}
}

func (e JobEventCancelled) JobName() string {
	return e.name
}

func (e JobEventCancelled) Timestamp() time.Time {
	return e.timestamp
}

func NewJobEventCompleted(name string, timestamp time.Time) JobEventCompleted {
	return JobEventCompleted{name: name, timestamp: timestamp}
}

func (e JobEventCompleted) JobName() string {
	return e.name
}

func (e JobEventCompleted) Timestamp() time.Time {
	return e.timestamp
}

type JobEventListener interface {
	On(context.Context, JobEvent)
}

type JobEventDispatcher struct {
	listeners []JobEventListener
}

func NewJobEventDispatcher() *JobEventDispatcher {
	return &JobEventDispatcher{
		listeners: make([]JobEventListener, 0),
	}
}

func (d *JobEventDispatcher) Add(listener JobEventListener) {
	d.listeners = append(d.listeners, listener)
}

func (d *JobEventDispatcher) On(ctx context.Context, event JobEvent) {
	for _, listener := range d.listeners {
		listener.On(ctx, event)
	}
}
