package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"scheduler/shared"
	"strings"
)

func main() {
	store := NewStore()

	http.HandleFunc("/tasks/", store.TaskHandler)
	http.ListenAndServe(":8080", nil)
}

type Store struct {
	tasks map[string]*shared.Task
}

func NewStore() *Store {
	return &Store{tasks: make(map[string]*shared.Task)}
}

func (store *Store) TaskHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	log.Println("Query", r.Method, r.URL.Path, pathParts)

	w.Header().Set("Content-Type", "application/json")

	if len(pathParts) == 3 && pathParts[2] != "" {
		taskId := pathParts[2]
		if r.Method == "GET" {
			log.Println("Get task %s", taskId)

			task, ok := store.tasks[taskId]

			if !ok {
				http.NotFound(w, r)
				return
			}

			write(w, task)
			return
		} else if r.Method == "PATCH" {
			log.Println("Starting", taskId)

			// TODO read attribute and merge
			task, ok := store.tasks[taskId]

			if !ok {
				http.NotFound(w, r)
				return
			}

			task.Status = shared.Running

			write(w, task)
			return
		}
	} else if r.Method == "POST" {

		decoder := json.NewDecoder(r.Body)
		var task *shared.Task
		decoder.Decode(&task)

		task.Uuid = uuid()
		store.tasks[task.Uuid] = task
		task.Status = shared.Pending

		log.Println("Saving %s %+v", uuid, task)

		write(w, task)
		return

	} else if r.Method == "GET" {
		log.Println("All tasks")

		encoded, _ := json.Marshal(store.tasks)
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
