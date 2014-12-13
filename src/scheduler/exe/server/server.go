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
	scheduler := server.NewScheduler()
	http.HandleFunc("/tasks/", scheduler.TaskHandler)
	http.ListenAndServe(":8080", nil)
}
