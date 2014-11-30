package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"scheduler/node"
	"scheduler/node/transport"
	"scheduler/shared"
	//	"strings"
	"testing"
	"time"
)

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

func TodoNode(t *testing.T) {

	store := NewStore()

	ts := httptest.NewServer(http.HandlerFunc(store.TaskHandler))
	defer ts.Close()

	task := shared.Task{Executable: "ls"}
	encoded, _ := json.Marshal(task)
	resp, _ := http.Post(ts.URL+"/tasks", "application/json", bytes.NewReader(encoded))

	location := resp.Header.Get("Location")
	//	uuid := strings.Split(location, "/")[2]

	transport := transport.HttpNodeTransport{Url: ts.URL}
	node := node.Start(1, transport)

	t.Log(node.Size)

	node.Run()

	timeout := make(chan bool, 1)
	ch := make(chan bool, 1)
	go func() {
		time.Sleep(2 * time.Second)
		timeout <- true
	}()
	go func() {
		//		for {
		taskResponse, _ := http.Get(ts.URL + location)
		decoder := json.NewDecoder(taskResponse.Body)
		var decodedTask shared.Task
		decoder.Decode(&decodedTask)
		if decodedTask.Status == shared.Finished {
			ch <- true
			//			break
		}
		//		}
	}()
	select {
	case <-ch:

	case <-timeout:
		t.Fatal("Timeout")
	}
}
