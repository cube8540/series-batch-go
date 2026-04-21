package book

import "strings"

type Site string

func NewSite(s string) Site {
	return Site(s)
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
	ISBN string
	Name string
}

func PrepareSeriesNameForSearch(name string) string {
	return strings.ReplaceAll(name, " ", "")
}
