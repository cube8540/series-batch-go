package book

import (
	"errors"
	"strings"
)

type Site string

const (
	SiteNLGO   Site = "NLGO"
	SiteNaver  Site = "NAVER"
	SiteAladin Site = "ALADIN"
	SiteKyobo  Site = "KYOBO"
)

func NewSite(s string) Site {
	switch strings.ToLower(s) {
	case "nlgo":
		return SiteNLGO
	case "naver":
		return SiteNaver
	case "aladin":
		return SiteAladin
	case "kyobo":
		return SiteKyobo
	default:
		return Site(s)
	}
}

type Original map[Site]map[string]any

func NewOriginal() Original {
	return make(map[Site]map[string]any)
}

type Book struct {
	ID     uint
	ISBN   string
	Title  string
	Series *Series

	OriginalData Original
}

type Series struct {
	ID   uint
	ISBN *string
	Name string
}

func PrepareSeriesNameForSearch(name string) string {
	return strings.ReplaceAll(name, " ", "")
}

type OriginalKey string

type OriginalDictionary map[OriginalKey]string

const (
	OriginalKeyTitle = OriginalKey("title")

	OriginalKeyISBN       = OriginalKey("isbn")
	OriginalKeySeriesISBN = OriginalKey("series_isbn")

	OriginalKeyPrice       = OriginalKey("price")
	OriginalKeyDescription = OriginalKey("description")

	OriginalKeySeriesList = "mapper"
)

type OriginalKeyMapper struct {
	dict OriginalDictionary
}

func OriginalKeyMapping(site Site) (*OriginalKeyMapper, error) {
	switch site {
	case SiteNLGO:
		return &OriginalKeyMapper{dictionaryNLGO()}, nil
	case SiteNaver:
		return &OriginalKeyMapper{dictionaryNaver()}, nil
	case SiteAladin:
		return &OriginalKeyMapper{dictionaryAladin()}, nil
	case SiteKyobo:
		return &OriginalKeyMapper{dictionaryKyobo()}, nil
	default:
		return nil, errors.New("unknown site")
	}
}

func (m *OriginalKeyMapper) Retrieve(o map[string]any, k OriginalKey) (any, bool) {
	if v, ok := o[m.dict[k]]; ok {
		return v, true
	}
	return nil, false
}

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
