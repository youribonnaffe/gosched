package server

import (
	"encoding/json"
	"log"
	"net/http"
	"scheduler/shared"
	"strconv"
	"strings"
)

// TODO move blocking logic to scheduler?
var nodePollingClients chan chan bool = make(chan chan bool)

func (scheduler *Scheduler) TaskHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	log.Println("Query", r.Method, r.URL.Path, pathParts)

	w.Header().Set("Content-Type", "application/json")

	taskId := isTaskSpecific(pathParts)
	if taskId != "" {
		if r.Method == "GET" {
			log.Println("Get task", taskId)

			task, found := scheduler.GetTask(taskId)

			if !found {
				http.NotFound(w, r)
				return
			}

			if len(pathParts) >= 4 && pathParts[3] == "output" {
				tailing := r.URL.Query().Get("tail")
				log.Println("Get task output tailing is", tailing)

				if tailing == "true" {
					fromLine, err := strconv.Atoi(r.URL.Query().Get("fromLine"))

					if err != nil {
						log.Println(err)
						http.Error(w, "fromLine is malformed", 400)
						return
					}

					taskOutput, err := scheduler.TailTaskOutput(taskId, fromLine)

					if err != nil {
						log.Println(err)
						http.Error(w, "Cannot get output", 500)
						return
					}

					log.Println(fromLine, len(taskOutput)-1, taskOutput[fromLine:len(taskOutput)])
					encoded, _ := json.Marshal(taskOutput[fromLine:len(taskOutput)])
					w.Write(encoded)
					return
				}
				encoded, _ := json.Marshal(task.Output)
				w.Write(encoded)
				return
			}

			write(w, &task)
			log.Println("Done Get task", taskId)

			return
		} else if r.Method == "PATCH" {
			log.Println("Updating", taskId)

			_, found := scheduler.GetTask(taskId)

			if !found {
				http.NotFound(w, r)
				return
			}

			if len(pathParts) >= 4 && pathParts[3] == "output" {
				log.Println("Adding task output")

				var output string
				decoder := json.NewDecoder(r.Body)
				err := decoder.Decode(&output)
				if err != nil {
					http.Error(w, "Task state is malformed", http.StatusInternalServerError)
					return
				} else {
					scheduler.AddOutputToTask(taskId, output)
					return
				}

			}

			newState := shared.Task{}
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&newState)

			if err != nil {
				http.Error(w, "Task state is malformed", http.StatusInternalServerError)
				return
			} else {
				log.Println("Update", newState)
				task, _ := scheduler.UpdateTask(newState)

				if err != nil {
					http.Error(w, "Task is already running", http.StatusInternalServerError)
					return
				}
				write(w, task)
				return
			}
		}
	} else if r.Method == "POST" {

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		var newTask shared.Task
		decoder.Decode(&newTask)

		task := scheduler.CreateTask(newTask)

		log.Println("Saving", task)

		unblockPolling()

		w.Header().Set("Location", "/tasks/"+task.Uuid)
		w.WriteHeader(http.StatusCreated)
		write(w, &task)
		return

	} else if r.Method == "GET" { // eventually consistent?

		polling := r.URL.Query().Get("polling")
		status := r.URL.Query().Get("status")
		log.Println("All tasks", status)

		if polling == "true" {
			tasks := scheduler.ListTasks(status)

			if len(tasks) > 0 {
				encoded, _ := json.Marshal(tasks)
				w.Write(encoded)
				return
			}

			blockPolling(func() {
				tasks := scheduler.ListTasks(status)

				encoded, _ := json.Marshal(tasks)
				w.Write(encoded)
				return
			})
		} else {
			tasks := scheduler.ListTasks(status)
			encoded, _ := json.Marshal(tasks)
			w.Write(encoded)
			return
		}
	}

	http.NotFound(w, r)
}

func unblockPolling() {
Loop:
	for {
		select {
		case pollingClient := <-nodePollingClients:
			pollingClient <- true
		default:
			break Loop
		}
	}
}

func blockPolling(function func()) {
	pollingClient := make(chan bool)

	select {
	case nodePollingClients <- pollingClient:
		<-pollingClient

		function()
	}
}

func isTaskSpecific(pathParts []string) string {
	if len(pathParts) >= 3 && pathParts[2] != "" {
		return pathParts[2]
	} else {
		return ""
	}
}

func write(w http.ResponseWriter, task *shared.Task) {
	encoded, _ := json.Marshal(task)
	w.Write(encoded)
}
