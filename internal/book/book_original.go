package book

import "errors"

type Original map[Site]map[string]any

func NewOriginal() Original {
	return make(map[Site]map[string]any)
}

type (
	OriginalKey        string
	OriginalDictionary map[OriginalKey]string
)

const (
	OriginalKeyTitle = OriginalKey("title")

	OriginalKeyISBN       = OriginalKey("isbn")
	OriginalKeySeriesISBN = OriginalKey("series_isbn")

	OriginalKeyPrice       = OriginalKey("price")
	OriginalKeyDescription = OriginalKey("description")

	OriginalKeySeriesList = "mapper"
)

func dictionaryNLGO() OriginalDictionary {
	m := make(OriginalDictionary)
	m[OriginalKeyTitle] = "title"
	m[OriginalKeySeriesISBN] = "set_isbn"
	return m
}

func dictionaryNaver() OriginalDictionary {
	m := make(OriginalDictionary)
	m[OriginalKeyTitle] = "title"
	m[OriginalKeyDescription] = "description"
	return m
}

func dictionaryAladin() OriginalDictionary {
	m := make(OriginalDictionary)
	m[OriginalKeyTitle] = "title"
	m[OriginalKeyDescription] = "description"
	return m
}

func dictionaryKyobo() OriginalDictionary {
	m := make(OriginalDictionary)
	m[OriginalKeyTitle] = "title"
	m[OriginalKeyDescription] = "prod_description"
	return m
}

type OriginalKeyMapper struct {
	dict OriginalDictionary
}

func NewOriginalKeyMapper(site Site) (OriginalKeyMapper, error) {
	switch site {
	case SiteNLGO:
		return OriginalKeyMapper{dictionaryNLGO()}, nil
	case SiteNaver:
		return OriginalKeyMapper{dictionaryNaver()}, nil
	case SiteAladin:
		return OriginalKeyMapper{dictionaryAladin()}, nil
	case SiteKyobo:
		return OriginalKeyMapper{dictionaryKyobo()}, nil
	default:
		return OriginalKeyMapper{}, errors.New("unknown site")
	}
}

func (m *OriginalKeyMapper) Get(o map[string]any, k OriginalKey) (any, bool) {
	if v, ok := o[m.dict[k]]; ok {
		return v, true
	}
	return nil, false
}
