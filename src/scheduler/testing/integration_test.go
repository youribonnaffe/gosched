package testing

import (
	"net/http"
	"net/http/httptest"
	"scheduler/client"
	"scheduler/node"
	"scheduler/node/transport"
	"scheduler/server"
	"scheduler/shared"
	"testing"
	"time"
)

const WAIT_TIMEOUT = 2 * time.Second

func TestIntegration_CreateTaskAndList(t *testing.T) {

	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("true")

	if err != nil {
		t.Error(err)
	}

	// list all tasks
	tasks, err := client.GetTasks()

	if err != nil {
		t.Error(err)
	}

	if tasks[0].Executable != "true" {
		t.Error("Expected a task that runs true")
	}

	if tasks[0].Status != shared.Pending {
		t.Error("Expected a pending task")
	}

	// list created task
	task, err := client.GetTask(createdTask.Uuid)

	if err != nil {
		t.Error(err)
	}

	if task.Executable != "true" {
		t.Error("Expected a task that runs true")
	}

	if task.Status != shared.Pending {
		t.Error("Expected a pending task")
	}
}

func TestIntegration_CreateTaskAndExecute(t *testing.T) {

	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("hostname")

	if err != nil {
		t.Error(err)
	}

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	node.Run()

	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)

	if finishedTask.Output == "" {
		t.Error("Output expected")
	}
}

func TestIntegration_CreateTaskAlreadyRunningNode(t *testing.T) {

	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	go func() {
		node.Run()
	}()

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("hostname")

	if err != nil {
		t.Error(err)
	}

	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)

	if finishedTask.Output == "" {
		t.Error("Output expected")
	}
}

func waitUntilTaskFinished(t *testing.T, client client.Client, uuid string) *shared.Task {
	timeout := make(chan bool, 1)
	ch := make(chan *shared.Task, 1)
	go func() {
		time.Sleep(WAIT_TIMEOUT)
		timeout <- true
	}()
	go func() {
		for {
			task, err := client.GetTask(uuid)

			if err != nil {
				t.Error(err)
			}

			if task.Status == shared.Finished {
				ch <- task
				break
			}
		}
	}()
	select {
	case task := <-ch:
		return task
	case <-timeout:
		t.Error("Timeout")
		return nil
	}
}
