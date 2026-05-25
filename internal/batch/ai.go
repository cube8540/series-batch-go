package batch

import (
	"context"
)

type AI interface {
	RequestSeriesNormalizeBatch(ctx context.Context, displayName string, requests []SeriesNormalizeRequest) (string, error)
	GetSeriesNormalizeBatch(ctx context.Context, jobName string) (Status, []SeriesNormalizeBatchResult, error)
}

type SeriesNormalizeRequest struct {
	Title    string
	SaleInfo []*SiteSaleInfo
}

type SiteSaleInfo struct {
	Site   string
	Title  string
	Price  uint
	Desc   *string
	Series []string
}

type SeriesNormalizeBatchResult struct {
	Key      string
	Response SeriesNormalizeResponse
}

type SeriesNormalizeResponse struct {
	Title string
	Noise []*SeriesNormalizeNoise
}

type SeriesNormalizeNoise struct {
	Text   string
	Reason string
}
