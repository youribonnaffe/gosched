package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"scheduler/shared"
	"strings"
)

type Store struct {
	Tasks map[string]*shared.Task
}

func NewStore() *Store {
	return &Store{Tasks: make(map[string]*shared.Task)}
}

func (store *Store) TaskHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	log.Println("Query", r.Method, r.URL.Path, pathParts)

	w.Header().Set("Content-Type", "application/json")

	if len(pathParts) == 3 && pathParts[2] != "" {
		taskId := pathParts[2]
		if r.Method == "GET" {
			log.Println("Get task", taskId)

			task, ok := store.Tasks[taskId]

			if !ok {
				http.NotFound(w, r)
				return
			}

			write(w, task)
			return
		} else if r.Method == "PATCH" {
			log.Println("Starting", taskId)

			// TODO read attribute and merge
			task, ok := store.Tasks[taskId]

			if !ok {
				http.NotFound(w, r)
				return
			}

			task.Status = shared.Running

			newState := shared.Task{}
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&newState)
			if err != nil {
				log.Println(err)
			} else {
				task.Status = newState.Status
				task.Output = newState.Output
				log.Println("Update", newState)
			}

			write(w, task)
			return
		}
	} else if r.Method == "POST" {

		decoder := json.NewDecoder(r.Body)
		var task *shared.Task
		decoder.Decode(&task)

		task.Uuid = uuid()
		store.Tasks[task.Uuid] = task
		task.Status = shared.Pending

		log.Println("Saving", uuid, task)

		w.Header().Set("Location", "/tasks/"+task.Uuid)
		w.WriteHeader(http.StatusCreated)
		write(w, task)
		return

	} else if r.Method == "GET" {

		status := r.URL.Query().Get("status")
		log.Println("All tasks", status)
		v := make([]*shared.Task, 0, len(store.Tasks))

		for _, value := range store.Tasks {
			if status == "" || value.Status == status {
				v = append(v, value)
			}
		}
		encoded, _ := json.Marshal(v)
		w.Write(encoded)
		return
	}
}

func write(w http.ResponseWriter, task *shared.Task) {
	encoded, _ := json.Marshal(task)
	w.Write(encoded)
}

func uuid() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}
