package shared

type Task struct {
	Executable string
	Status     string
	Uuid       string
	Output     string
}

const (
	Pending  string = "pending"
	Running  string = "running"
	Finished string = "finished"
)
