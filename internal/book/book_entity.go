package book

type bookEntity struct {
	ID    uint
	ISBN  string
	Title string

	SeriesID *uint
	Series   *seriesEntity
}

func (e *bookEntity) domain() *Book {
	b := &Book{
		ID:           e.ID,
		ISBN:         e.ISBN,
		Title:        e.Title,
		OriginalData: NewOriginal(),
	}

	if e.Series != nil {
		b.Series = e.Series.domain()
	}

	return b
}

func (e *bookEntity) TableName() string {
	return "books.book"
}

type originalEntity struct {
	BookID     uint
	Site       string
	OriginData map[string]any `gorm:"serializer:json;type:json"`
}

func (e originalEntity) TableName() string {
	return "books.book_origin_data"
}

type seriesEntity struct {
	ID           uint
	ISBN         *string
	Name         string
	NameFullText string
}

func (e *seriesEntity) domain() *Series {
	return &Series{
		ID:   e.ID,
		ISBN: e.ISBN,
		Name: e.Name,
	}
}

func (e *seriesEntity) TableName() string {
	return "books.series"
}
