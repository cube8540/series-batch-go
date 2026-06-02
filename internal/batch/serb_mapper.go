package batch

import (
	"context"
	"errors"
	"fmt"
	"series-batch-go/internal/book"
	"series-batch-go/internal/pkg/collections"
	"series-batch-go/internal/pkg/structs"
	"series-batch-go/internal/schedule"
)

type SeriesMappingReader struct {
	ai AI

	repository       book.ConditionalRepository
	seriesRepository book.ConditionalSeriesRepository
	batchRepository  Repository

	limit int
}

func NewSeriesMapperReader(ai AI, repository book.ConditionalRepository, seriesRepository book.ConditionalSeriesRepository, batchRepository Repository) *SeriesMappingReader {
	return &SeriesMappingReader{
		ai:               ai,
		repository:       repository,
		seriesRepository: seriesRepository,
		batchRepository:  batchRepository,
		limit:            100,
	}
}

func (reader *SeriesMappingReader) SetLimit(limit int) {
	reader.limit = limit
}

type Mapped struct {
	batch *Batch
	pair  *structs.Pair[*book.Book, string]
}

func (reader *SeriesMappingReader) Read(ctx context.Context, _ schedule.JobParameter) ([]*Mapped, error) {
	batches := reader.batchRepository.GetByStatus(ctx, reader.limit, StatusDone)

	var mapped []*Mapped
	for _, batch := range batches {
		_, batchResults, err := reader.ai.GetSeriesNormalizeBatch(ctx, batch.ExternalID)
		if err != nil {
			return nil, err
		}

		books := reader.repository.Get(ctx, batch.GetTargetID()...)
		bookMap := collections.Map(books, func(b *book.Book) uint {
			return b.ID
		})

		for _, batchResult := range batchResults {
			idx := collections.Find(batch.Targets, func(e Target) bool {
				return e.RequestID == batchResult.Key
			})
			if idx > -1 {
				bok, normalizedTitle := bookMap[batch.Targets[idx].BookID], batchResult.Response.Title
				mapped = append(mapped, &Mapped{batch: batch, pair: structs.NewPair(bok, normalizedTitle)})
			} else {
				return nil, fmt.Errorf("(%d) batch result not found: %s", batch.ID, batchResult.Key)
			}
		}
	}

	return mapped, nil
}

type SeriesMappingWriter struct {
	batchRepository Repository

	repository       book.ConditionalRepository
	seriesRepository book.ConditionalSeriesRepository
}

func NewSeriesMappingWriter(batchRepository Repository, repository book.ConditionalRepository, seriesRepository book.ConditionalSeriesRepository) *SeriesMappingWriter {
	return &SeriesMappingWriter{
		batchRepository:  batchRepository,
		repository:       repository,
		seriesRepository: seriesRepository,
	}
}

func (writer *SeriesMappingWriter) Write(ctx context.Context, items []*Mapped) error {
	for _, item := range items {
		bok, norTitle := item.pair.First, item.pair.Second

		var seriesISBN string
		var matchedSeries *book.Series

		mapper, _ := book.NewOriginalKeyMapper(book.SiteNLGO)
		if isbnAny, ok := mapper.Get(bok.OriginalData[book.SiteNLGO], book.OriginalKeySeriesISBN); ok {
			if series := writer.seriesRepository.Get(ctx, isbnAny.(string)); len(series) > 0 {
				matchedSeries = series[0]
			}
		}

		if matchedSeries == nil {
			series := writer.seriesRepository.TitleFullTextSearch(ctx, norTitle)
			if len(series) > 0 {
				matchedSeries = series[0].First
			}
		}

		if matchedSeries == nil {
			if seriesISBN != "" {
				matchedSeries = &book.Series{ISBN: &seriesISBN, Name: norTitle}
			} else {
				matchedSeries = &book.Series{Name: norTitle}
			}
			if _, err := writer.seriesRepository.Save(ctx, []*book.Series{matchedSeries}); err != nil {
				return err
			}
		}

		bok.Series = matchedSeries
		if err := writer.repository.Update(ctx, bok); err != nil {
			return err
		}
		item.batch.Status = StatusCompleted
	}

	batchSet := make(map[uint]*Batch)
	for _, item := range items {
		if _, ok := batchSet[item.batch.ID]; !ok {
			batchSet[item.batch.ID] = item.batch
		}
	}

	var errArr []error
	for _, batch := range batchSet {
		if err := writer.batchRepository.Update(ctx, batch); err != nil {
			errArr = append(errArr, err)
		}
	}
	if len(errArr) > 0 {
		return errors.Join(errArr...)
	}
	return nil
}
