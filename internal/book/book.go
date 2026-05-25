package book

import "strings"

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
