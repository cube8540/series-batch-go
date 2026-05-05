package mapper

import (
	"context"
	"errors"
	"series-batch-go/internal/batch"
	"series-batch-go/internal/book"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/gemini"
	"series-batch-go/internal/pkg/collections"
	"series-batch-go/internal/pkg/llm"
	"series-batch-go/internal/pkg/structs"
)

type Reader struct {
	geminiClient *gemini.Client

	batchRepository  *batch.Repository
	bookRepository   *book.Repository
	seriesRepository *book.SeriesRepository

	limit int
}

func NewReader(geminiClient *gemini.Client, batchRepository *batch.Repository, bookRepository *book.Repository, seriesRepository *book.SeriesRepository) *Reader {
	return &Reader{
		geminiClient:     geminiClient,
		batchRepository:  batchRepository,
		bookRepository:   bookRepository,
		seriesRepository: seriesRepository,
		limit:            100,
	}
}

func (r *Reader) SetLimit(limit int) {
	r.limit = limit
}

type ReadItem struct {
	BatchID uint

	// Pair 읽어온 도서의 아이디와 제목 묶음
	Pair *structs.Pair[uint, string]
}

func (r *Reader) Read(ctx context.Context, _ map[string]string) []*ReadItem {
	batches := r.batchRepository.FindBatchByStatus(ctx, r.limit, llm.JobStatusDone)

	var items []*ReadItem
	for _, btc := range batches {
		_, batchResults, err := r.geminiClient.GetSeriesNormalizeBatch(ctx, btc.ExternalID)
		if err != nil {
			log.Sugared().Errorf("error occurred when getting mapper normalize batch: %v", err)
		}

		for _, batchResult := range batchResults {
			idx := collections.Find(btc.Targets, func(e batch.Target) bool {
				return e.RequestID == batchResult.Key
			})
			if idx > -1 {
				items = append(items, &ReadItem{
					BatchID: btc.ID,
					Pair:    structs.NewPair(btc.Targets[idx].BookID, batchResult.Response.Title),
				})
			} else {
				log.Sugared().Errorf("(%d) batch result not found: %s", btc.ID, batchResult.Key)
			}
		}
	}
	return items
}

type Processor struct {
	bookRepository   *book.Repository
	seriesRepository *book.SeriesRepository
}

func NewProcessor(bookRepository *book.Repository, seriesRepository *book.SeriesRepository) *Processor {
	return &Processor{
		bookRepository:   bookRepository,
		seriesRepository: seriesRepository,
	}
}

type ProcessItem struct {
	BatchID uint
	Pair    *structs.Pair[*book.Book, *book.Series]
}

func (p *Processor) Process(ctx context.Context, input *ReadItem) *ProcessItem {
	b := p.bookRepository.FindByID(ctx, input.Pair.First)[0]
	mapper, _ := book.OriginalKeyMapping(book.SiteNLGO)

	var isbn string
	if isbnAny, ok := mapper.Retrieve(b.OriginalData[book.SiteNLGO], book.OriginalKeySeriesISBN); ok {
		isbn = isbnAny.(string)
		if series := p.seriesRepository.FindSeriesByISBN(ctx, isbn); len(series) > 0 {
			return &ProcessItem{BatchID: input.BatchID, Pair: structs.NewPair(b, series[0])}
		}
	}

	series := p.seriesRepository.FindSeriesByFullTextSearch(ctx, input.Pair.Second)
	if len(series) > 0 {
		return &ProcessItem{BatchID: input.BatchID, Pair: structs.NewPair(b, series[0].First)}
	}

	if isbn != "" {
		return &ProcessItem{BatchID: input.BatchID, Pair: structs.NewPair(b, &book.Series{ISBN: &isbn, Name: input.Pair.Second})}
	}
	return &ProcessItem{BatchID: input.BatchID, Pair: structs.NewPair(b, &book.Series{Name: input.Pair.Second})}
}

type Writer struct {
	batchRepository *batch.Repository

	bookRepository   *book.Repository
	seriesRepository *book.SeriesRepository
}

func NewWriter(batchRepository *batch.Repository, bookRepository *book.Repository, seriesRepository *book.SeriesRepository) *Writer {
	return &Writer{
		batchRepository:  batchRepository,
		bookRepository:   bookRepository,
		seriesRepository: seriesRepository,
	}
}

func (w *Writer) Write(ctx context.Context, items []*ProcessItem) error {
	batchID := make(map[uint]struct{})
	for _, i := range items {
		b, s := i.Pair.First, i.Pair.Second
		if s.ID == 0 {
			if ss, err := w.seriesRepository.SaveSeries(ctx, []*book.Series{s}); err == nil {
				s = ss[0]
			} else {
				return err
			}
		}
		if err := w.bookRepository.UpdateBookSeries(ctx, b.ID, s.ID); err != nil {
			return err
		}
		batchID[i.BatchID] = struct{}{}
	}

	var errs []error
	for id, _ := range batchID {
		if err := w.batchRepository.UpdateBatchStatus(ctx, id, llm.JobStatusCompleted); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
