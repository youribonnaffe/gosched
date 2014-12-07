package main

import (
	"log"
	"net/http"
	"runtime"
	"scheduler/server"
)

func main() {
	runtime.GOMAXPROCS(8)
	log.SetFlags(log.Lmicroseconds)
	store := server.NewStore()

	http.HandleFunc("/tasks/", store.TaskHandler)
	http.ListenAndServe(":8080", nil)
}
