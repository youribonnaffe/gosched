package shared

import (
	"errors"
)

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

func (task *Task) ChangeStatus(newStatus string) error {
	if task.Status == Running && newStatus == Running {
		return errors.New("Task already running!")
	} else {
		task.Status = newStatus
		return nil
	}
}
