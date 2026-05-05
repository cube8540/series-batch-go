package repository

import (
	"context"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/pkg/structs"

	"gorm.io/gorm"
)

func FindBookByID(ctx context.Context, db *gorm.DB, ID ...uint) []*BookEntity {
	books, err := gorm.G[*BookEntity](db.Unscoped()).Where("id IN (?)", ID).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding books(%v): %v", ID, err)
	}
	return books
}

func UpdateBookSeries(ctx context.Context, db *gorm.DB, ID uint, seriesID uint) error {
	_, err := gorm.G[BookEntity](db).Where("id = ?", ID).Updates(ctx, BookEntity{SeriesID: &seriesID})
	return err
}

func FindUnorganizedBooks(ctx context.Context, db *gorm.DB, limit int) []*BookEntity {
	books, err := gorm.G[*BookEntity](db.Unscoped()).Where("series_id is null").Order("id desc").Limit(limit).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding unorgized books: %v", err)
	}
	return books
}

func FindBookOriginalDataByBookID(ctx context.Context, db *gorm.DB, bookID ...uint) []*BookOriginalDataEntity {
	entities, err := gorm.G[*BookOriginalDataEntity](db).Where("book_id IN (?)", bookID).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding book original data(%v): %v", bookID, err)
	}
	return entities
}

func FindSeriesByISBN(ctx context.Context, db *gorm.DB, ISBN ...string) []*SeriesEntity {
	entities, err := gorm.G[*SeriesEntity](db).Where("isbn IN (?)", ISBN).Find(ctx)
	if err != nil {
		log.Sugared().Errorf("error occurred when finding mapper(%v): %v", ISBN, err)
	}
	return entities
}

func FindSeriesByFullTextSearch(ctx context.Context, db *gorm.DB, name string) []*structs.Pair[*SeriesEntity, float32] {
	type seriesFullTextSearch struct {
		ID           uint
		ISBN         *string
		Name         string
		NameFullText string
		Score        float32
	}

	var res []*seriesFullTextSearch
	selected := db.WithContext(ctx).Model(&SeriesEntity{}).
		Select("id, isbn, name, name_full_text, bigm_similarity(name_full_text, ?) as score", name).
		Where("name_full_text =% ?", name).
		Order("score DESC").
		Find(&res)

	if selected.Error != nil {
		log.Sugared().Errorf("error occurred when finding mapper by full text search: %v", selected.Error)
	}

	var results []*structs.Pair[*SeriesEntity, float32]
	for _, r := range res {
		entity := &SeriesEntity{
			ID:           r.ID,
			ISBN:         r.ISBN,
			Name:         r.Name,
			NameFullText: r.NameFullText,
		}
		results = append(results, structs.NewPair(entity, r.Score))
	}

	return results
}

func SaveSeries(ctx context.Context, db *gorm.DB, series []*SeriesEntity) error {
	result := db.WithContext(ctx).Create(&series)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
