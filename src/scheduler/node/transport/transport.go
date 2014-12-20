package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"scheduler/shared"
)

type Protocol interface {
	ListTasks(status string) ([]shared.Task, error)
	Update(shared.Task) error
	AddOutputToTask(uuid string, output string) error
}

type HttpNodeTransport struct {
	Url string
}

func (t HttpNodeTransport) ListTasks(status string) (tasks []shared.Task, err error) {

	r, err := http.Get(t.Url + "/tasks/?polling=true&status=" + status)
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

	r, err := client.Do(req)

	if err == nil {
		_, err = io.Copy(ioutil.Discard, r.Body)
		defer r.Body.Close()
	} else {
		return err
	}
	if r.StatusCode != 200 {
		bs, _ := ioutil.ReadAll(r.Body)
		return errors.New(string(bs))
	}
	return err
}

func (t HttpNodeTransport) AddOutputToTask(uuid string, output string) error {
	encoded, _ := json.Marshal(output)

	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", t.Url+"/tasks"+"/"+uuid+"/output", bytes.NewReader(encoded))

	r, err := client.Do(req)

	if err == nil {
		_, err = io.Copy(ioutil.Discard, r.Body)
		defer r.Body.Close()
	} else {
		return err
	}
	if r.StatusCode != 200 {
		bs, _ := ioutil.ReadAll(r.Body)
		return errors.New(string(bs))
	}
	return err
}
