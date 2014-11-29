package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scheduler/shared"
	"time"
)

func main() {

	// list all jobs
	// take a pending one
	// run it
	// mark it as finished
	// and again

	workers := make(chan *shared.Task)
	go worker(workers)
	go worker(workers)

	ticker := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff

				r, _ := http.Get("http://localhost:8080/tasks")

				decoder := json.NewDecoder(r.Body)

				tasks := make(map[string]*shared.Task)

				decoder.Decode(&tasks)

				for id, task := range tasks {
					if task.Status == shared.Pending {

						log.Println("starting on", id)
						workers <- task
						log.Println("started on", id)
						break
					}
				}

				defer r.Body.Close()

			case <-quit:
				log.Println("done")
				ticker.Stop()
				return
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Println("sig", sig)
		}
	}()
	select {}
	log.Println("end")
}

func worker(tasks <-chan *shared.Task) {
	for t := range tasks {
		log.Println("working on", t.Uuid)
		patch("http://localhost:8080/tasks/" + t.Uuid)
		// TODO check it is ok to start working on it
		time.Sleep(10 * time.Second)
		log.Println("done working on", t.Uuid)
	}
}

func patch(path string) (resp *http.Response, err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", path, nil)
	return client.Do(req)
}
