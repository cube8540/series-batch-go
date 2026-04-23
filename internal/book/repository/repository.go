package repository

import (
	"context"
	"gorm.io/gorm"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/pkg/structs"
)

var (
	dummyBooks            = BookEntity{}
	dummyBookOriginalData = BookOriginalDataEntity{}

	dummySeries = SeriesEntity{}
)

func FindBookByID(ctx context.Context, db *gorm.DB, ID ...uint) []*BookEntity {
	var books []*BookEntity
	if err := db.WithContext(ctx).Where(dummyBooks.TableName()+".ID IN (?)", ID).Find(&books).Error; err != nil {
		log.Sugared().Errorf("error occurred when finding books(%v): %v", ID, err)
	}
	return books
}

func FindUnorganizedBooks(ctx context.Context, db *gorm.DB, limit int) []*BookEntity {
	var books []*BookEntity
	if err := db.WithContext(ctx).Where(dummyBooks.TableName() + ".series_id IS NULL").Order("id desc").Limit(limit).Find(&books).Error; err != nil {
		log.Sugared().Errorf("error occurred when finding unorgized books: %v", err)
	}
	return books
}

func FindBookOriginalDataByBookID(ctx context.Context, db *gorm.DB, bookID ...uint) []*BookOriginalDataEntity {
	var entities []*BookOriginalDataEntity
	if err := db.WithContext(ctx).Debug().Where(dummyBookOriginalData.TableName()+".book_id IN (?)", bookID).Find(&entities).Error; err != nil {
		log.Sugared().Errorf("error occurred when finding book original data(%v): %v", bookID, err)
	}
	return entities
}

func FindSeriesByISBN(ctx context.Context, db *gorm.DB, ISBN ...string) []*SeriesEntity {
	var entities []*SeriesEntity
	if err := db.WithContext(ctx).Where(dummySeries.TableName()+".ISBN IN (?)", ISBN).Find(&entities).Error; err != nil {
		log.Sugared().Errorf("error occurred when finding series(%v): %v", ISBN, err)
	}
	return entities
}

func FindSeriesByFullTextSearch(ctx context.Context, db *gorm.DB, name string) []*structs.Pair[*SeriesEntity, float32] {
	type seriesFullTextSearch struct {
		ID           uint
		ISBN         string
		Name         string
		NameFullText string
		Score        float32
	}

	var res []*seriesFullTextSearch
	err := db.WithContext(ctx).Model(&seriesFullTextSearch{}).
		Select("id, isbn, name, name_full_text, bigm_similarity(name, ?) as score", name).
		Where("name_full_text LIKE =% ?", name).
		Order("score DESC").
		Find(&res)

	if err != nil {
		log.Sugared().Errorf("error occurred when finding series(%v): %v", name, err)
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
