package batch

import (
	"context"
	"series-batch-go/internal/book"
	"series-batch-go/internal/schedule"
	"strconv"
	"time"
)

type SeriesClassifierReader struct {
	repo  book.ConditionalRepository
	limit int
}

func NewSeriesClassifierReader(repo book.ConditionalRepository) *SeriesClassifierReader {
	return &SeriesClassifierReader{
		repo:  repo,
		limit: 100,
	}
}

func (reader SeriesClassifierReader) SetLimit(limit int) {
	reader.limit = limit
}

func (reader SeriesClassifierReader) Read(ctx context.Context, _ schedule.JobParameter) ([]*book.Book, error) {
	return reader.repo.GetUnorganized(ctx, reader.limit), nil
}

type SeriesClassifierWriter struct {
	ai              AI
	batchRepository Repository

	GenerateDisplayName func() string
}

func NewSeriesClassifierWriter(ai AI, batchRepository Repository) *SeriesClassifierWriter {
	return &SeriesClassifierWriter{
		ai:              ai,
		batchRepository: batchRepository,
	}
}

func (writer *SeriesClassifierWriter) Write(ctx context.Context, items []*book.Book) error {
	var requests []SeriesNormalizeRequest
	for _, item := range items {
		request := SeriesNormalizeRequest{
			Title: item.Title,
		}

		var sale []*SiteSaleInfo
		for site, dict := range item.OriginalData {
			mapper, _ := book.NewOriginalKeyMapper(site)
			saleInfo := &SiteSaleInfo{
				Site: string(site),
			}
			if title, ok := mapper.Get(dict, book.OriginalKeyTitle); ok {
				saleInfo.Title = title.(string)
			}
			if desc, ok := mapper.Get(dict, book.OriginalKeyDescription); ok {
				d := desc.(string)
				saleInfo.Desc = &d
			}
			sale = append(sale, saleInfo)
		}
		request.SaleInfo = sale
		requests = append(requests, request)
	}

	batchName, err := writer.ai.RequestSeriesNormalizeBatch(ctx, writer.GenerateDisplayName(), requests)
	if err != nil {
		return err
	}

	batch := &Batch{
		ExternalID:   batchName,
		Type:         TypeSeriesNormalization,
		Status:       StatusPending,
		RegisteredAt: time.Now(),
	}

	for i, item := range items {
		batch.Targets = append(batch.Targets, Target{
			BookID:    item.ID,
			RequestID: strconv.Itoa(i),
		})
	}

	_, err = writer.batchRepository.Save(ctx, []*Batch{batch})
	return err
}
