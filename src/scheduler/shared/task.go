package shared

type Task struct {
	Executable string
	Status     Status
	Uuid       string
}

type Status string

const (
	Pending  Status = "pending"
	Running  Status = "running"
	Finished Status = "finished"
)
