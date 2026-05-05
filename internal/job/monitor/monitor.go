package monitor

import (
	"context"
	"errors"
	"series-batch-go/internal/batch"
	"series-batch-go/internal/gemini"
	"series-batch-go/internal/pkg/llm"
	"strings"
)

type ReadItem struct {
	BatchID uint
	ExID    string

	State llm.JobStatus
}

type ProcessItem struct {
	BatchID uint
	ExID    string

	Before llm.JobStatus
	After  llm.JobStatus
}

type Reader struct {
	batchRepository *batch.Repository
	limit           int
}

func NewReader(batchRepository *batch.Repository) *Reader {
	return &Reader{
		batchRepository: batchRepository,
		limit:           100,
	}
}

func (r *Reader) SetLimit(limit int) {
	r.limit = limit
}

func (r *Reader) Read(ctx context.Context, p map[string]string) []*ReadItem {
	var states []llm.JobStatus
	if s, ok := p["state"]; ok {
		s = strings.ReplaceAll(s, " ", "")
		for _, s = range strings.Split(s, ",") {
			states = append(states, llm.JobStatus(s))
		}
	} else {
		states = append(states, llm.JobStatusRunning, llm.JobStatusPending)
	}

	jobs := r.batchRepository.FindBatchByStatus(ctx, r.limit, states...)
	var items []*ReadItem
	for _, job := range jobs {
		items = append(items, &ReadItem{
			BatchID: job.ID,
			ExID:    job.ExternalID,
			State:   job.State,
		})
	}
	return items
}

type Processor struct {
	geminiClient *gemini.Client
}

func NewProcessor(geminiClient *gemini.Client) *Processor {
	return &Processor{geminiClient: geminiClient}
}

func (p *Processor) Process(ctx context.Context, item *ReadItem) *ProcessItem {
	status, _, _ := p.geminiClient.GetSeriesNormalizeBatch(ctx, item.ExID)

	return &ProcessItem{
		BatchID: item.BatchID,
		ExID:    item.ExID,
		Before:  item.State,
		After:   status,
	}
}

type Writer struct {
	batchRepository *batch.Repository
}

func NewWriter(batchRepository *batch.Repository) *Writer {
	return &Writer{
		batchRepository: batchRepository,
	}
}

func (w *Writer) Write(ctx context.Context, items []*ProcessItem) error {
	var errs []error
	for _, item := range items {
		if item.Before != item.After {
			if err := w.batchRepository.UpdateBatchStatus(ctx, item.BatchID, item.After); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}
