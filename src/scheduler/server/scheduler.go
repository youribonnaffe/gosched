package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"scheduler/shared"
	"sync"
	"time"
)

type Scheduler struct {
	lock  sync.RWMutex
	tasks map[string]*LockedTask
}

type LockedTask struct {
	lock sync.Mutex
	task *shared.Task
}

func NewScheduler() *Scheduler {
	return &Scheduler{tasks: make(map[string]*LockedTask)}
}

func (scheduler *Scheduler) CreateTask(newTask shared.Task) shared.Task {
	task := shared.Task{}
	task.Uuid = uuid()
	task.Status = shared.Pending
	task.Executable = newTask.Executable
	task.SubmittedTime = time.Now()

	scheduler.lock.Lock() // TODO on read too
	scheduler.tasks[task.Uuid] = &LockedTask{task: &task}
	scheduler.lock.Unlock()

	return task
}

func (scheduler *Scheduler) GetTask(uuid string) (shared.Task, bool) {
	lockedTask, found := scheduler.tasks[uuid]
	if !found {
		return shared.Task{}, found
	}
	return *lockedTask.task, found
}

func (scheduler *Scheduler) UpdateTask(newState shared.Task) (*shared.Task, error) {
	// update and do locking
	lockedTask, found := scheduler.tasks[newState.Uuid]

	if !found {
		return nil, errors.New("Task not found")
	}

	lockedTask.lock.Lock()

	err := changeStatus(lockedTask.task, newState.Status)

	if err != nil {
		lockedTask.lock.Unlock()
		return nil, err
	}

	lockedTask.task.Output = newState.Output
	lockedTask.task.StartTime = newState.StartTime
	lockedTask.task.ExecutionDuration = newState.ExecutionDuration

	lockedTask.lock.Unlock()
	return lockedTask.task, nil
}

func (scheduler *Scheduler) ListTasks(status string) []shared.Task {
	v := make([]shared.Task, 0, len(scheduler.tasks))

	for _, value := range scheduler.tasks {
		if status == "" || value.task.Status == status {
			v = append(v, *value.task)
		}
	}

	return v
}

func changeStatus(task *shared.Task, newStatus string) error {
	if task.Status != shared.Pending && newStatus == shared.Running {
		return errors.New("Task already running or already finished!")
	} else {
		task.Status = newStatus
		return nil
	}
}

func uuid() (uuid string) { // TODO test collision?
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}
