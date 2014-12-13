package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"scheduler/node"
	"scheduler/node/transport"
	"scheduler/shared"
	//	"strings"
	"bytes"
	"scheduler/client"
	"testing"
	"time"
)

const WAIT_TIMEOUT = 2 * time.Second

func TestSubmitATask(t *testing.T) {
	store := NewStore()
	task := shared.Task{Executable: "ls"}

	encoded, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", "http://localhost:8080/tasks", bytes.NewReader(encoded))

	w := httptest.NewRecorder()
	store.TaskHandler(w, req)

	if w.Code != 201 {
		t.Fatal("A POST should return 201 upon success")
	}

	location := w.HeaderMap.Get("Location")

	listTaskReq, _ := http.NewRequest("GET", "http://localhost:8080"+location, nil)
	listTaskWriter := httptest.NewRecorder()
	store.TaskHandler(listTaskWriter, listTaskReq)

	if listTaskWriter.Code != 200 {
		t.Fatal("A GET should return 200 upon success")
	}

	decoder := json.NewDecoder(listTaskWriter.Body)
	var decodedTask shared.Task
	decoder.Decode(&decodedTask)

	if decodedTask.Executable != "ls" {
		t.Fatal("Wrong executable saved")
	}

	listTasksReq, _ := http.NewRequest("GET", "http://localhost:8080/tasks", nil)
	listTasksWriter := httptest.NewRecorder()
	store.TaskHandler(listTasksWriter, listTasksReq)

	if listTasksWriter.Code != 200 {
		t.Fatal("A GET should return 200 upon success")
	}

	var decodedTasks []shared.Task
	json.NewDecoder(listTasksWriter.Body).Decode(&decodedTasks)

	if decodedTasks[0].Executable != "ls" {
		t.Fatal("Wrong executable saved", decodedTasks)
	}

}

func TestIntegration_CreateTaskAndList(t *testing.T) {

	store := NewStore()
	server := httptest.NewServer(http.HandlerFunc(store.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("true")

	if err != nil {
		t.Fatal(err)
	}

	// list all tasks
	tasks, err := client.GetTasks()

	if err != nil {
		t.Fatal(err)
	}

	if tasks[0].Executable != "true" {
		t.Fatal("Expected a task that runs true")
	}

	if tasks[0].Status != shared.Pending {
		t.Fatal("Expected a pending task")
	}

	// list created task
	task, err := client.GetTask(createdTask.Uuid)

	if err != nil {
		t.Fatal(err)
	}

	if task.Executable != "true" {
		t.Fatal("Expected a task that runs true")
	}

	if task.Status != shared.Pending {
		t.Fatal("Expected a pending task")
	}
}

func TestIntegration_CreateTaskAndExecute(t *testing.T) {

	store := NewStore()
	server := httptest.NewServer(http.HandlerFunc(store.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("hostname")

	if err != nil {
		t.Fatal(err)
	}

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	node.Run()

	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)

	if finishedTask.Output == "" {
		t.Fatal("Output expected")
	}
}

func TestIntegration_CreateTaskAlreadyRunningNode(t *testing.T) {

	store := NewStore()
	server := httptest.NewServer(http.HandlerFunc(store.TaskHandler))
	defer server.Close()

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	go func() {
		node.Run()
	}()

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("hostname")

	if err != nil {
		t.Fatal(err)
	}

	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)

	if finishedTask.Output == "" {
		t.Fatal("Output expected")
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
				t.Fatal(err)
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
		t.Fatal("Timeout")
		return nil
	}
}

// give goconvey a shot? testing with real commands (node, server, client)
// when a task is submitted, it can be listed in all tasks
// when a task is submitted, it can queried by id
// when a node is started and a task submitted, it is executed (and with several nodes too)
// when a task submitted and a node started, it is executed (and with several nodes too)
