package shared

import (
	"fmt"
	"time"
)

type Task struct {
	Executable        string
	Status            string
	Uuid              string
	Output            []string
	SubmittedTime     time.Time
	StartTime         time.Time
	ExecutionDuration time.Duration
}

const (
	Pending  string = "pending"
	Running  string = "running"
	Finished string = "finished"
)

func (t *Task) String() string {
	return fmt.Sprintf("%s %s %s %d",
		t.Uuid, t.Executable, t.Status,
		t.ExecutionDuration.Nanoseconds())
}

type BySubmittedTime []*Task

func (a BySubmittedTime) Len() int           { return len(a) }
func (a BySubmittedTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySubmittedTime) Less(i, j int) bool { return a[i].SubmittedTime.Before(a[j].SubmittedTime) }
