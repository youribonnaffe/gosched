package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"scheduler/shared"
)

type Client struct {
	Url    string
	client http.Client
}

func (c *Client) Execute(executable string) (*shared.Task, error) {

	task := shared.Task{Executable: executable}

	encoded, _ := json.Marshal(task)
	req, err := c.client.Post(c.Url+"/tasks/", "application/json", bytes.NewReader(encoded))

	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	if req.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(req.Body)
		return nil, errors.New("Could not create task: " + string(body) + req.Status)
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&task)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *Client) GetTasks() (tasks []*shared.Task, err error) {

	r, err := http.Get(c.Url + "/tasks/")
	if err == nil {
		defer r.Body.Close()
	} else {
		return nil, err
	}

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&tasks)

	return tasks, nil
}

func (c *Client) GetTask(uuid string) (task *shared.Task, err error) {

	r, err := http.Get(c.Url + "/tasks/" + uuid)
	if err == nil {
		defer r.Body.Close()
	} else {
		return nil, err
	}

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&task)

	return task, nil
}

func (c *Client) Tail(uuid string) (lines []string, err error) {

	r, err := http.Get(c.Url + "/tasks/" + uuid + "/output?tail=true")
	if err == nil {
		defer r.Body.Close()
	} else {
		return
	}

	if r.StatusCode != http.StatusOK {
		err = errors.New("Task not found")
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&lines)

	return
}
