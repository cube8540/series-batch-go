package book

import (
	"context"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/pkg/structs"
	"slices"

	"gorm.io/gorm"
)

const DefaultGormBatchSize = 100

type GormRepository struct {
	db        *gorm.DB
	BatchSize int
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db, BatchSize: DefaultGormBatchSize}
}

func (repo *GormRepository) Get(ctx context.Context, ID ...uint) []*Book {
	var books []*Book
	for chunk := range slices.Chunk(ID, repo.BatchSize) {
		entities, err := gorm.G[*bookEntity](repo.db.Unscoped()).Where("id IN (?)", chunk).Find(ctx)
		if err != nil {
			log.Sugared().Errorf("error occurred when finding books(%v): %v", chunk, err)
		} else {
			books = append(books, repo.fillOriginalData(ctx, entities)...)
		}
	}
	return books
}

func (repo *GormRepository) Update(ctx context.Context, book *Book) error {
	entity := bookEntity{
		ID:    book.ID,
		ISBN:  book.ISBN,
		Title: book.Title,
	}

	if book.Series != nil {
		entity.SeriesID = &book.Series.ID
	}

	_, err := gorm.G[bookEntity](repo.db).Select("*").Updates(ctx, entity)
	return err
}

func (repo *GormRepository) GetUnorganized(ctx context.Context, limit int) []*Book {
	type raw struct {
		ID    uint
		ISBN  string
		Title string
	}
	sql := `
		select id, isbn, title
		from books.book book
		where not exists (select 1 from books.series_batch_target where book_id = book.id) and
			  series_id is null
		order by id desc
		limit ?
	`
	raws, err := gorm.G[raw](repo.db).Raw(sql, limit).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding unorganized books: %v", err)
	}

	var entities []*bookEntity
	for _, row := range raws {
		entity := bookEntity{
			ID:    row.ID,
			ISBN:  row.ISBN,
			Title: row.Title,
		}
		entities = append(entities, &entity)
	}
	return repo.fillOriginalData(ctx, entities)
}

func (repo *GormRepository) fillOriginalData(ctx context.Context, entities []*bookEntity) []*Book {
	var (
		ID    []uint
		books []*Book
		m     = make(map[uint]*Book)
	)

	for _, entity := range entities {
		b := entity.domain()

		m[entity.ID] = b
		ID = append(ID, entity.ID)
		books = append(books, b)
	}

	originals, err := gorm.G[*originalEntity](repo.db).Where("book_id IN (?)", ID).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding book original data(%v): %v", ID, err)
	}
	for _, o := range originals {
		m[o.BookID].OriginalData[NewSite(o.Site)] = o.OriginData
	}

	return books
}

type SeriesGormRepository struct {
	db *gorm.DB
}

func NewSeriesGormRepository(db *gorm.DB) *SeriesGormRepository {
	return &SeriesGormRepository{db: db}
}

func (repo *SeriesGormRepository) Get(ctx context.Context, ISBN ...string) []*Series {
	entities, err := gorm.G[*seriesEntity](repo.db).Where("isbn IN (?)", ISBN).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding series(%v): %v", ISBN, err)
	}
	var series []*Series
	for _, entity := range entities {
		series = append(series, entity.domain())
	}
	return series
}

func (repo *SeriesGormRepository) Save(ctx context.Context, series []*Series) ([]*Series, error) {
	var entities []*seriesEntity
	for _, ser := range series {
		entities = append(entities, &seriesEntity{
			ISBN:         ser.ISBN,
			Name:         ser.Name,
			NameFullText: PrepareSeriesNameForSearch(ser.Name),
		})
	}

	if result := repo.db.WithContext(ctx).Create(&entities); result.Error != nil {
		return nil, result.Error
	}

	for i, entity := range entities {
		series[i].ID = entity.ID
	}
	return series, nil
}

func (repo *SeriesGormRepository) TitleFullTextSearch(ctx context.Context, query string) []*structs.Pair[*Series, float32] {
	type seriesFullTextSearch struct {
		ID           uint
		ISBN         *string
		Name         string
		NameFullText string
		Score        float32
	}

	var resp []*seriesFullTextSearch
	selected := repo.db.WithContext(ctx).Model(&seriesEntity{}).
		Select("id, isbn, name, name_full_text, bigm_similarity(name_full_text, ?) as score", PrepareSeriesNameForSearch(query)).
		Where("name_full_text =% ?", PrepareSeriesNameForSearch(query)).
		Order("score DESC").
		Find(&resp)

	if selected.Error != nil {
		log.Sugared().Errorf("error occurred when finding mapper by full text search: %v", selected.Error)
	}

	var results []*structs.Pair[*Series, float32]
	for _, ser := range resp {
		series := &Series{
			ID:   ser.ID,
			ISBN: ser.ISBN,
			Name: ser.Name,
		}
		results = append(results, structs.NewPair(series, ser.Score))
	}

	return results
}
