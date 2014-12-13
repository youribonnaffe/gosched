package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"scheduler/shared"
	"testing"
)

func TestSubmitATask(t *testing.T) {
	scheduler := NewScheduler()
	task := shared.Task{Executable: "ls"}

	encoded, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", "http://localhost:8080/tasks", bytes.NewReader(encoded))

	w := httptest.NewRecorder()
	scheduler.TaskHandler(w, req)

	if w.Code != 201 {
		t.Fatal("A POST should return 201 upon success")
	}

	location := w.HeaderMap.Get("Location")

	listTaskReq, _ := http.NewRequest("GET", "http://localhost:8080"+location, nil)
	listTaskWriter := httptest.NewRecorder()
	scheduler.TaskHandler(listTaskWriter, listTaskReq)

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
	scheduler.TaskHandler(listTasksWriter, listTasksReq)

	if listTasksWriter.Code != 200 {
		t.Fatal("A GET should return 200 upon success")
	}

	var decodedTasks []shared.Task
	json.NewDecoder(listTasksWriter.Body).Decode(&decodedTasks)

	if decodedTasks[0].Executable != "ls" {
		t.Fatal("Wrong executable saved", decodedTasks)
	}

}
