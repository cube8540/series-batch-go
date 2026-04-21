package book

import (
	"context"
	"gorm.io/gorm"
	"series-batch-go/internal/book/repository"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindUnorganizedBooks(ctx context.Context, limit int) []*Book {
	entities := repository.FindUnorganizedBooks(ctx, r.db, limit)

	var (
		ID    []uint
		books []*Book
		m     = make(map[uint]*Book)
	)

	for _, entity := range entities {
		b := ConvertBookEntityToDomain(entity)

		m[entity.ID] = b
		ID = append(ID, entity.ID)
		books = append(books, b)
	}

	originals := repository.FindBookOriginalDataByBookID(ctx, r.db, ID...)
	for _, o := range originals {
		m[o.BookID].OriginalData[NewSite(o.Site)] = o.OriginData
	}

	return books
}

type SeriesRepository struct {
	db *gorm.DB
}

func NewSeriesRepository(db *gorm.DB) *SeriesRepository {
	return &SeriesRepository{db: db}
}

func (r *SeriesRepository) FindSeriesByISBN(ctx context.Context, ISBN ...string) []*Series {
	entities := repository.FindSeriesByISBN(ctx, r.db, ISBN...)

	var series []*Series
	for _, entity := range entities {
		series = append(series, ConvertSeriesEntityToDomain(entity))
	}
	return series
}

func (r *SeriesRepository) FindSeriesByFullTextSearch(ctx context.Context, name string) []*Series {
	entities := repository.FindSeriesByFullTextSearch(ctx, r.db, PrepareSeriesNameForSearch(name))
	var series []*Series
	for _, entity := range entities {
		series = append(series, ConvertSeriesEntityToDomain(entity))
	}
	return series
}

func ConvertBookEntityToDomain(entity *repository.BookEntity) *Book {
	domain := &Book{
		ID:           entity.ID,
		ISBN:         entity.ISBN,
		Title:        entity.Title,
		OriginalData: NewOriginal(),
	}

	if entity.Series != nil {
		domain.Series = ConvertSeriesEntityToDomain(entity.Series)
	}

	return domain
}

func ConvertSeriesEntityToDomain(entity *repository.SeriesEntity) *Series {
	return &Series{
		ID:   entity.ID,
		ISBN: entity.ISBN,
		Name: entity.Name,
	}
}
