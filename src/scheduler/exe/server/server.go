package main

import (
	"net/http"
	"scheduler/server"
)

func main() {
	store := server.NewStore()

	http.HandleFunc("/tasks/", store.TaskHandler)
	http.ListenAndServe(":8080", nil)
}
