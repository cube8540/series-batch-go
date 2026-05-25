package schedule

type JobStatus string

const (
	JobStatusIdle    JobStatus = "IDLE"
	JobStatusRunning JobStatus = "RUNNING"
)

type Instance struct {
	Name  string
	State JobStatus
}

func NewInstance(name string, status JobStatus) Instance {
	return Instance{Name: name, State: status}
}
