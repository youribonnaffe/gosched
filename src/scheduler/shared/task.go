package shared

import (
	"time"
)

type Task struct {
	Executable        string
	Status            string
	Uuid              string
	Output            string
	SubmittedTime     time.Time
	StartTime         time.Time
	ExecutionDuration time.Duration
}

const (
	Pending  string = "pending"
	Running  string = "running"
	Finished string = "finished"
)
