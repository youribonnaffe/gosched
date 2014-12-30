package server

import (
	"scheduler/shared"
	"testing"
)

func TestTailingOneLine(t *testing.T) {

	scheduler := NewScheduler()
	task := shared.Task{Executable: "ls"}

	task = scheduler.CreateTask(task)

	scheduler.AddOutputToTask(task.Uuid, "heyhey")
	scheduler.UpdateTask(shared.Task{Uuid: task.Uuid, Status: shared.Finished})

	output, err := scheduler.TailTaskOutput(task.Uuid, 0)

	if err != nil {
		t.Error(err)
	}

	if output[0] != "heyhey" {
		t.Error("Output expected")
	}
}

func TestTailingNoLines(t *testing.T) {

	scheduler := NewScheduler()
	task := shared.Task{Executable: "ls"}

	task = scheduler.CreateTask(task)

	scheduler.UpdateTask(shared.Task{Uuid: task.Uuid, Status: shared.Finished})

	output, err := scheduler.TailTaskOutput(task.Uuid, 0)

	if err != nil {
		t.Error(err)
	}

	if len(output) != 0 {
		t.Error("No output expected")
	}
}

func TestTailingInvalidNumberOfLine(t *testing.T) {

	scheduler := NewScheduler()
	task := shared.Task{Executable: "ls"}

	task = scheduler.CreateTask(task)

	scheduler.AddOutputToTask(task.Uuid, "heyhey")
	scheduler.UpdateTask(shared.Task{Uuid: task.Uuid, Status: shared.Finished})

	_, err := scheduler.TailTaskOutput(task.Uuid, 1)

	if err == nil {
		t.Error("Error expected")
	}
}
