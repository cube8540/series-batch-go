package book

import (
	"context"
	"series-batch-go/internal/pkg/structs"
)

type Repository interface {
	Get(ctx context.Context, ID ...uint) []*Book
	Update(ctx context.Context, book *Book) error
}

type ConditionalRepository interface {
	Repository
	GetUnorganized(ctx context.Context, limit int) []*Book
}

type SeriesRepository interface {
	Get(ctx context.Context, ISBN ...string) []*Series
	Save(ctx context.Context, series []*Series) ([]*Series, error)
}

type ConditionalSeriesRepository interface {
	SeriesRepository
	TitleFullTextSearch(ctx context.Context, query string) []*structs.Pair[*Series, float32]
}
