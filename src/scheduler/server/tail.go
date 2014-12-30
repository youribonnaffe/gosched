package server

import (
	"errors"
	"scheduler/shared"
)

func (scheduler *Scheduler) TailTaskOutput(uuid string, fromLine int) (output []string, err error) {
	lockedTask, found := scheduler.tasks[uuid]

	if !found {
		return nil, errors.New("Task not found")

	}

	if lockedTask.task.Status != shared.Finished { // TODO lock?
		client := make(chan struct{})

		select {
		case lockedTask.taskOutputClients <- client:
			select {
			case <-client:
				task := lockedTask.task
				return task.Output[fromLine:len(task.Output)], nil

			}
		}
	} else {
		task := lockedTask.task
		if len(task.Output) > 0 && fromLine >= len(task.Output) {
			return nil, errors.New("Not enough lines in task output")
		}
		return task.Output[fromLine:len(task.Output)], nil
	}
}

func unblockOutputClients(lockedTask *lockedTask) {
Loop:
	for {
		select {
		case client := <-lockedTask.taskOutputClients:
			client <- struct{}{}
		default:
			break Loop
		}
	}
}

func closeTaskOutputClients(lockedTask *lockedTask) {
	if lockedTask.task.Status == shared.Finished {
	Loop:
		for {
			select {
			case pollingClient := <-lockedTask.taskOutputClients:
				close(pollingClient)
			default:
				break Loop
			}
		}
	}
}
