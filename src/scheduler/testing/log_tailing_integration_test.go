package testing

import (
	"net/http"
	"net/http/httptest"
	"scheduler/client"
	"scheduler/node"
	"scheduler/node/transport"
	"scheduler/server"
	"testing"
)

func TestIntegration_Tail_TaskNotStarted_TaskFinished(t *testing.T) {

	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("hostname")

	if err != nil {
		t.Error(err)
	}

	tailedLine := make(chan string)

	go func() {
		lines, err := client.Tail(createdTask.Uuid)
		if err != nil {
			t.Fatal(err)
		}

		tailedLine <- lines[0]
	}()

	go func() {
		node.Run()
	}()

	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)

	line := <-tailedLine

	if err != nil {
		t.Fatal(err)
	}

	if line == "" {
		t.Fatalf("Log expected")
	}

	lines, err := client.Tail(finishedTask.Uuid)

	if err != nil {
		t.Fatal(err)
	}

	if lines[0] == "" {
		t.Fatalf("Log expected")
	}

}

//func TestIntegration_TaskRunning(t *testing.T) {
//
//	scheduler := server.NewScheduler()
//	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
//	defer server.Close()
//
//	transport := transport.HttpNodeTransport{Url: server.URL}
//	node := node.Start(1, transport)
//
//	client := client.Client{Url: server.URL}
//	createdTask, err := client.Execute("hostname")
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	emptyLine, err := client.Tail(createdTask.Uuid)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if emptyLine != "" {
//		t.Fatalf("No log expected")
//	}
//
//	go func() {
//		node.Run()
//	}()
//
//	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)
//
//	line, err := client.Tail(finishedTask.Uuid)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if line == "" {
//		t.Fatalf("Log expected")
//	}
//
//}
