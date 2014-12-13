package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"scheduler/shared"
	"strings"
	"sync"
)

type Store struct {
	lock  sync.RWMutex
	Tasks map[string]*LockedTask
}

type LockedTask struct {
	lock sync.Mutex
	Task *shared.Task
}

var pollingClients chan chan bool = make(chan chan bool)

func NewStore() *Store {
	return &Store{Tasks: make(map[string]*LockedTask)}
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

			write(w, task.Task)
			return
		} else if r.Method == "PATCH" {
			log.Println("Updating", taskId)

			// TODO read attribute and merge

			// lock task

			task, ok := store.Tasks[taskId]

			task.lock.Lock()
			if !ok {
				http.NotFound(w, r)
				return
			}

			newState := shared.Task{}
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&newState)
			if err != nil {
				log.Println(err)
			} else {
				err := task.Task.ChangeStatus(newState.Status)
				if err != nil {
					http.Error(w, "Task already running", http.StatusInternalServerError)
					task.lock.Unlock()
					return
				}
				task.Task.Output = newState.Output
				log.Println("Update", newState)
			}

			task.lock.Unlock()

			write(w, task.Task)
			log.Println("Done Updating", taskId)
			return
		}
	} else if r.Method == "POST" {

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		var task *shared.Task
		decoder.Decode(&task)

		task.Uuid = uuid()
		task.Status = shared.Pending

		store.lock.Lock() // TODO on read too
		store.Tasks[task.Uuid] = &LockedTask{Task: task}
		store.lock.Unlock()

		log.Println("Saving", task)

	Loop:
		for {
			select {
			case pollingClient := <-pollingClients:
				pollingClient <- true
			default:
				break Loop
			}
		}

		w.Header().Set("Location", "/tasks/"+task.Uuid)
		w.WriteHeader(http.StatusCreated)
		write(w, task)
		return

	} else if r.Method == "GET" { // eventually consistent?

		polling := r.URL.Query().Get("polling")
		status := r.URL.Query().Get("status")
		log.Println("All tasks", status)

		if polling == "true" {
			v := make([]*shared.Task, 0, len(store.Tasks))

			for _, value := range store.Tasks {
				if status == "" || value.Task.Status == status {
					v = append(v, value.Task)
				}
			}

			if len(v) > 0 {
				encoded, _ := json.Marshal(v)
				w.Write(encoded)
				return
			}

			pollingClient := make(chan bool)

			select {
			case pollingClients <- pollingClient:
				<-pollingClient

				v := make([]*shared.Task, 0, len(store.Tasks))

				for _, value := range store.Tasks {
					if status == "" || value.Task.Status == status {
						v = append(v, value.Task)
					}
				}

				encoded, _ := json.Marshal(v)
				w.Write(encoded)
				return
			}
		}

		v := make([]*shared.Task, 0, len(store.Tasks))

		for _, value := range store.Tasks {
			if status == "" || value.Task.Status == status {
				v = append(v, value.Task)
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
