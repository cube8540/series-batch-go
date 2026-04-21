package repository

type BookEntity struct {
	ID    uint
	ISBN  string
	Title string

	SeriesID *uint
	Series   *SeriesEntity
}

func (e *BookEntity) TableName() string {
	return "books.book"
}

type BookOriginalDataEntity struct {
	BookID     uint
	Site       string
	OriginData map[string]any `gorm:"serializer:json;type:json"`
}

func (e BookOriginalDataEntity) TableName() string {
	return "books.book_origin_data"
}

type SeriesEntity struct {
	ID           uint
	ISBN         string
	Name         string
	NameFullText string
}

func (e *SeriesEntity) TableName() string {
	return "books.series"
}
