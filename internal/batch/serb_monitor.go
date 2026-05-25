package batch

import (
	"context"
	"errors"
	"series-batch-go/internal/schedule"
	"strings"
)

type SeriesMonitorReader struct {
	ai AI

	batchRepository Repository
	limit           int
}

func NewSeriesMonitorReader(ai AI, batchRepository Repository) *SeriesMonitorReader {
	return &SeriesMonitorReader{
		ai:              ai,
		batchRepository: batchRepository,
		limit:           100,
	}
}

func (reader *SeriesMonitorReader) SetLimit(limit int) {
	reader.limit = limit
}

func (reader *SeriesMonitorReader) Read(ctx context.Context, params schedule.JobParameter) ([]*Batch, error) {
	var status []Status
	if state, ok := params["state"]; ok {
		state = strings.ReplaceAll(state, " ", "")
		for _, s := range strings.Split(state, ",") {
			status = append(status, Status(s))
		}
	} else {
		status = []Status{StatusRunning, StatusPending}
	}

	batches := reader.batchRepository.GetByStatus(ctx, reader.limit, status...)

	var changedBatches []*Batch
	for _, batch := range batches {
		state, _, err := reader.ai.GetSeriesNormalizeBatch(ctx, batch.ExternalID)
		if err != nil {
			return nil, err
		}
		if state != batch.Status {
			batch.Status = state
			changedBatches = append(changedBatches, batch)
		}
	}

	return changedBatches, nil
}

type SeriesMonitorWriter struct {
	batchRepository Repository
}

func NewSeriesMonitorWriter(batchRepository Repository) *SeriesMonitorWriter {
	return &SeriesMonitorWriter{
		batchRepository: batchRepository,
	}
}

func (writer *SeriesMonitorWriter) Write(ctx context.Context, items []*Batch) error {
	var errs []error
	for _, item := range items {
		if err := writer.batchRepository.Update(ctx, item); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
