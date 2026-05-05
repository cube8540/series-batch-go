package classifier

import (
	"context"
	"series-batch-go/internal/batch"
	"series-batch-go/internal/book"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/gemini"
	"series-batch-go/internal/pkg/llm"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Reader struct {
	batchRepository *batch.Repository
	bookRepository  *book.Repository

	limit int
}

func NewReader(batchRepository *batch.Repository, bookRepository *book.Repository) *Reader {
	return &Reader{
		batchRepository: batchRepository,
		bookRepository:  bookRepository,
		limit:           100,
	}
}

func (r *Reader) SetLimit(limit int) {
	r.limit = limit
}

func (r *Reader) Read(ctx context.Context, _ map[string]string) []*book.Book {
	batches := r.batchRepository.FindBatchByStatus(ctx, 1, llm.JobStatusPending, llm.JobStatusRunning)
	if len(batches) > 0 {
		log.Sugared().Info("이미 실행중인 배치가 존재하여 새로운 배치를 만들 수 없습니다.")
		return []*book.Book{}
	}

	return r.bookRepository.FindUnorganizedBooks(ctx, r.limit)
}

type IdentifyProcessor struct{}

func (p *IdentifyProcessor) Process(_ context.Context, b *book.Book) *book.Book {
	return b
}

type Writer struct {
	geminiClient    *gemini.Client
	batchRepository *batch.Repository

	genDisplayName func() string
}

func NewWriter(geminiClient *gemini.Client, batchRepository *batch.Repository) *Writer {
	return &Writer{
		geminiClient:    geminiClient,
		batchRepository: batchRepository,
		genDisplayName: func() string {
			uid, _ := uuid.NewUUID()
			return uid.String()
		},
	}
}

func (w *Writer) SetGenDisplayName(f func() string) {
	w.genDisplayName = f
}

func (w *Writer) Write(ctx context.Context, books []*book.Book) error {
	batchName, err := w.geminiClient.RunSeriesNormalizeBatch(ctx, w.genDisplayName(), convertBooksToBatchRequest(books))
	if err != nil {
		return err
	}

	bch := &batch.Batch{
		ExternalID: batchName,

		Type:  batch.TypeSeriesNormalization,
		State: llm.JobStatusPending,

		RegisteredAt: time.Now(),
	}

	for i, b := range books {
		bch.Targets = append(bch.Targets, batch.Target{
			BookID:    b.ID,
			RequestID: strconv.Itoa(i),
		})
	}

	_, err = w.batchRepository.SaveBatch(ctx, []*batch.Batch{bch})
	return err
}

func convertBooksToBatchRequest(books []*book.Book) []*llm.SeriesNormalizeRequest {
	var requests []*llm.SeriesNormalizeRequest
	for _, b := range books {
		request := &llm.SeriesNormalizeRequest{Title: b.Title}

		var saleInfo []*llm.SiteSaleInfo
		for site, m := range b.OriginalData {
			mapper, _ := book.OriginalKeyMapping(site)

			si := &llm.SiteSaleInfo{}
			si.Site = string(site)

			if title, ok := mapper.Retrieve(m, book.OriginalKeyTitle); ok {
				si.Title = title.(string)
			}

			if desc, ok := mapper.Retrieve(m, book.OriginalKeyDescription); ok {
				d := desc.(string)
				si.Desc = &d
			}

			saleInfo = append(saleInfo, si)
		}

		request.SaleInfo = saleInfo
		requests = append(requests, request)
	}
	return requests
}
