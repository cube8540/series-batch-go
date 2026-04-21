package repository

import (
	"context"
	"gorm.io/gorm"
	"series-batch-go/internal/config/log"
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

func FindSeriesByFullTextSearch(ctx context.Context, db *gorm.DB, name string) []*SeriesEntity {
	var entities []*SeriesEntity
	if err := db.WithContext(ctx).Debug().Where("name_full_text LIKE ?", "%"+name+"%").Find(&entities).Error; err != nil {
		log.Sugared().Errorf("error occurred when finding series(%v): %v", name, err)
	}
	return entities
}
