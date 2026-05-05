package batch

import (
	"series-batch-go/internal/pkg/llm"
	"time"
)

type Type string

const (
	TypeSeriesNormalization Type = "SERIES_NORMALIZATION"
)

func NewType(s string) Type {
	switch s {
	case "SERIES_NORMALIZATION":
		return TypeSeriesNormalization
	default:
		return Type(s)
	}
}

type Batch struct {
	ID         uint
	ExternalID string

	Type  Type
	State llm.JobStatus

	RegisteredAt time.Time
	ModifiedAt   *time.Time

	Targets []Target
}

func NewSeriesNormalizationBatch(ID string, state llm.JobStatus) *Batch {
	return &Batch{
		ExternalID:   ID,
		Type:         TypeSeriesNormalization,
		State:        state,
		RegisteredAt: time.Now(),
	}
}

func (b *Batch) GetTargetID() []uint {
	var ID []uint
	for _, t := range b.Targets {
		ID = append(ID, t.BookID)
	}
	return ID
}

type Target struct {
	BookID    uint
	RequestID string
}
