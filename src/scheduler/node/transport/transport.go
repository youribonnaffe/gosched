package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"scheduler/shared"
)

type Protocol interface {
	ListTasks(status string) ([]shared.Task, error)
	Update(shared.Task) error
}

type HttpNodeTransport struct {
	Url string
}

func (t HttpNodeTransport) ListTasks(status string) (tasks []shared.Task, err error) {

	r, err := http.Get(t.Url + "/tasks?status=" + status)
	if err == nil {
		defer r.Body.Close()
	} else {
		return nil, err
	}

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&tasks)

	return tasks, nil
}

func (t HttpNodeTransport) Update(task shared.Task) error {
	encoded, _ := json.Marshal(task)

	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", t.Url+"/tasks"+"/"+task.Uuid, bytes.NewReader(encoded))

	resp, err := client.Do(req)

	if err == nil {
		_, err = io.Copy(ioutil.Discard, resp.Body)
		defer resp.Body.Close()
	}
	return err
}
