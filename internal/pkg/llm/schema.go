package llm

// JobStatus LLM 배치 잡 상태
type JobStatus string

const (
	// JobStatusUndefined 잡 상태를 확인할 수 없음을 나타낸다.
	JobStatusUndefined JobStatus = "UNDEFINED"

	// JobStatusPending 현재 잡이 실행 대기중임을 나타낸다.
	JobStatusPending JobStatus = "PENDING"

	// JobStatusRunning 현재 잡이 실행중임을 나타낸다.
	JobStatusRunning JobStatus = "RUNNING"

	// JobStatusCanceled 잡 실행이 취소 되었음을 나타낸다.
	JobStatusCanceled JobStatus = "CANCELED"

	// JobStatusFailed 잡이 실패했음을 나타낸다.
	JobStatusFailed JobStatus = "FAILED"

	// JobStatusDone 잡이 완료 되었음을 나타낸다.
	JobStatusDone JobStatus = "DONE"

	// JobStatusRetry 실패한 잡을 재실행중임을 나타낸다.
	JobStatusRetry JobStatus = "RETRY"

	// JobStatusCompleted 모든 작업이 완료되었음을 나타낸다.
	JobStatusCompleted JobStatus = "COMPLETED"
)

// SeriesNormalizeRequest 도서 시리즈 일반화 요청
type SeriesNormalizeRequest struct {
	// Title 일반화를 요청할 도서의 제목
	Title string

	// SaleInfo 사이트별 도서 판매 정보
	SaleInfo []SiteSaleInfo
}

// SiteSaleInfo 도서 판매 정보
type SiteSaleInfo struct {
	// Site 판매 사이트
	Site string

	// Title 도서 제목
	Title string

	// Price 판매가
	Price uint

	// Desc 판매처에서 입력한 도서 소개(혹은 설명)
	Desc *string

	// Series 판매처에서 정한 도서와 같은 시리즈 도서명
	Series []string
}

type SeriesNormalizeResponse struct {
	Title string
	Noise []*SeriesNormalizeNoise
}

type SeriesNormalizeNoise struct {
	Text   string
	Reason string
}
