package book

import (
	"context"
	"gorm.io/gorm"
	"series-batch-go/internal/book/repository"
	"series-batch-go/internal/pkg/structs"
	"slices"
)

type Repository struct {
	db *gorm.DB

	BatchSize int
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db, BatchSize: 100}
}

func (r *Repository) FindByID(ctx context.Context, ID ...uint) []*Book {
	var books []*Book
	for chunk := range slices.Chunk(ID, r.BatchSize) {
		entities := repository.FindBookByID(ctx, r.db, chunk...)
		b := r.fillOriginalData(ctx, entities)

		books = append(books, b...)
	}
	return books
}

func (r *Repository) FindUnorganizedBooks(ctx context.Context, limit int) []*Book {
	entities := repository.FindUnorganizedBooks(ctx, r.db, limit)
	return r.fillOriginalData(ctx, entities)
}

func (r *Repository) UpdateBookSeries(ctx context.Context, ID uint, seriesID uint) error {
	return repository.UpdateBookSeries(ctx, r.db, ID, seriesID)
}

func (r *Repository) fillOriginalData(ctx context.Context, entities []*repository.BookEntity) []*Book {
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

func (r *SeriesRepository) FindSeriesByFullTextSearch(ctx context.Context, name string) []*structs.Pair[*Series, float32] {
	search := repository.FindSeriesByFullTextSearch(ctx, r.db, PrepareSeriesNameForSearch(name))

	var result []*structs.Pair[*Series, float32]
	for _, r := range search {
		series := ConvertSeriesEntityToDomain(r.First)
		result = append(result, structs.NewPair(series, r.Second))
	}

	return result
}

func (r *SeriesRepository) SaveSeries(ctx context.Context, series []*Series) ([]*Series, error) {
	var entities []*repository.SeriesEntity
	for _, s := range series {
		entities = append(entities, &repository.SeriesEntity{
			ISBN:         s.ISBN,
			Name:         s.Name,
			NameFullText: PrepareSeriesNameForSearch(s.Name),
		})
	}
	if err := repository.SaveSeries(ctx, r.db, entities); err != nil {
		return nil, err
	}
	var result []*Series
	for _, e := range entities {
		result = append(result, ConvertSeriesEntityToDomain(e))
	}
	return result, nil
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
