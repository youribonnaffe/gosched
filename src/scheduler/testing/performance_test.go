package testing

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"scheduler/client"
	"scheduler/node"
	"scheduler/node/transport"
	"scheduler/server"
	"scheduler/shared"
	"testing"
)

func BenchmarkCreateTask(b *testing.B) {
	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}

	log.SetOutput(ioutil.Discard)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.Execute("true")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListTask(b *testing.B) {
	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}

	log.SetOutput(ioutil.Discard)

	client.Execute("true")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.GetTasks()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTaskExecution(b *testing.B) {
	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	client := client.Client{Url: server.URL}

	log.SetOutput(ioutil.Discard)
	log.SetFlags(log.Lmicroseconds)

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Execute("true")
		node.Run()

		ch := make(chan bool, 1)
		go func() {
			for {
				tasks, err := client.GetTasks()

				if err != nil {
					b.Error(err)
				}
				finished := true
				for _, task := range tasks {
					if task.Status != shared.Finished {
						finished = false
						break
					}
				}
				if finished {
					ch <- true
					return
				}
			}
		}()

	Loop:
		for {
			select {
			case <-ch:
				break Loop
			}
		}
		log.Println("Done running")

	}
}
