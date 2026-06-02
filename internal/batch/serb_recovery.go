package batch

import (
	"context"
	"series-batch-go/internal/book"
	"series-batch-go/internal/pkg/collections"
	"series-batch-go/internal/schedule"
)

const RecoveryBatchNames = "batch/schedule/shared/recovery_batches"

type RecoveryBatchReader struct {
	batchRepository Repository
	bookRepository  book.Repository
}

func NewRecoveryBatchReader(batchRepository Repository, bookRepository book.Repository) *RecoveryBatchReader {
	return &RecoveryBatchReader{
		batchRepository: batchRepository,
		bookRepository:  bookRepository,
	}
}

func (reader RecoveryBatchReader) Read(ctx context.Context, _ schedule.JobParameter) ([]*book.Book, error) {
	batches := reader.batchRepository.GetByStatus(ctx, 100, StatusFailed)

	targets := make(map[uint]any)
	var batchID []uint
	for _, batch := range batches {
		for _, target := range batch.Targets {
			targets[target.BookID] = nil
		}
		batchID = append(batchID, batch.ID)
	}

	if shared := schedule.GetSharedResource(ctx); shared != nil {
		shared[RecoveryBatchNames] = batchID
	}

	targetID := collections.MapToSlice(targets, func(k uint, _ any) uint {
		return k
	})
	return reader.bookRepository.Get(ctx, targetID...), nil
}

type RecoveryCompletedEventListener struct {
	batchRepository Repository
}

func NewRecoveryCompletedEventListener(batchRepository Repository) *RecoveryCompletedEventListener {
	return &RecoveryCompletedEventListener{
		batchRepository: batchRepository,
	}
}

func (listener RecoveryCompletedEventListener) On(ctx context.Context, event schedule.JobEvent) {
	if _, ok := event.(schedule.JobEventCompleted); ok {
		shared := schedule.GetSharedResource(ctx)
		batches := shared[RecoveryBatchNames].([]uint)

		recovered := listener.batchRepository.Get(ctx, batches...)
		for _, batch := range recovered {
			batch.Status = StatusCompleted
			_ = listener.batchRepository.Update(ctx, batch)
		}
	}
}
