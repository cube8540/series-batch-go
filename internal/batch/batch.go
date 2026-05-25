package batch

import (
	"time"
)

type (
	Type   string
	Status string
)

const (
	TypeSeriesNormalization Type = "SERIES_NORMALIZATION"

	StatusUndefined Status = "UNDEFINED"
	StatusPending   Status = "PENDING"
	StatusRunning   Status = "RUNNING"
	StatusCancelled Status = "CANCELED"
	StatusFailed    Status = "FAILED"
	StatusDone      Status = "DONE"
	StatusRetry     Status = "RETRY"
	StatusCompleted Status = "COMPLETED"
)

func NewType(s string) Type {
	switch s {
	case "SERIES_NORMALIZATION":
		return TypeSeriesNormalization
	default:
		return Type(s)
	}
}

func NewStatus(s string) Status {
	switch s {
	case "UNDEFINED":
		return StatusUndefined
	case "PENDING":
		return StatusPending
	case "RUNNING":
		return StatusRunning
	case "CANCELED":
		return StatusCancelled
	case "FAILED":
		return StatusFailed
	case "DONE":
		return StatusDone
	case "RETRY":
		return StatusRetry
	case "COMPLETED":
		return StatusCompleted
	default:
		return Status(s)
	}
}

type Batch struct {
	ID         uint
	ExternalID string

	Type   Type
	Status Status

	RegisteredAt time.Time
	ModifiedAt   *time.Time

	Targets []Target
}

type Target struct {
	BookID    uint
	RequestID string
}

func NewSeriesNormalizationBatch(ID string, status Status) *Batch {
	return &Batch{
		ExternalID:   ID,
		Type:         TypeSeriesNormalization,
		Status:       status,
		RegisteredAt: time.Now(),
	}
}

func (batch *Batch) GetTargetID() []uint {
	var ID []uint
	for _, t := range batch.Targets {
		ID = append(ID, t.BookID)
	}
	return ID
}

type Job struct {
	ID   uint
	Name string

	State     string
	UpdatedAt time.Time
}
