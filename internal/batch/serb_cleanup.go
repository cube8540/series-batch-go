package batch

import (
	"context"
	"series-batch-go/internal/schedule"
)

type CleanupBatchReader struct {
	batchRepository Repository
}

func NewCleanupBatchReader(batchRepository Repository) *CleanupBatchReader {
	return &CleanupBatchReader{
		batchRepository: batchRepository,
	}
}

func (reader CleanupBatchReader) Read(ctx context.Context, _ schedule.JobParameter) ([]*Batch, error) {
	return reader.batchRepository.GetByStatus(ctx, 100, StatusCompleted), nil
}

type CleanupBatchWriter struct {
	batchRepository Repository
}

func NewCleanupBatchWriter(batchRepository Repository) *CleanupBatchWriter {
	return &CleanupBatchWriter{
		batchRepository: batchRepository,
	}
}

func (writer CleanupBatchWriter) Write(ctx context.Context, batches []*Batch) error {
	return writer.batchRepository.Remove(ctx, batches)
}
