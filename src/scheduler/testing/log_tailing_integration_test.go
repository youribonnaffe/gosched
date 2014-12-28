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
		lines, err := client.Tail(createdTask.Uuid, 0)
		if err != nil {
			t.Error(err)
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

	lines, err := client.Tail(finishedTask.Uuid, 0)

	if err != nil {
		t.Fatal(err)
	}

	if lines[0] == "" {
		t.Fatalf("Log expected")
	}

}

func TestIntegration_TaskRunning(t *testing.T) {

	scheduler := server.NewScheduler()
	server := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))
	defer server.Close()

	transport := transport.HttpNodeTransport{Url: server.URL}
	node := node.Start(1, transport)

	client := client.Client{Url: server.URL}
	createdTask, err := client.Execute("/home/youri/dev/gosched/test.sh")

	if err != nil {
		t.Error(err)
	}

	go func() {
		fromLine := 0
		for {

			lines, err := client.Tail(createdTask.Uuid, fromLine)

			if err != nil {
				t.Error(err)
			}

			if len(lines) == 0 {
				return
			}

			if lines[0] == "" {
				t.Error("Log expected")
			}

			t.Log(lines[0])
			fromLine += len(lines)
		}
	}()

	go func() {
		node.Run()
	}()

	finishedTask := waitUntilTaskFinished(t, client, createdTask.Uuid)

	lines, err := client.Tail(finishedTask.Uuid, 0)

	if err != nil {
		t.Fatal(err)
	}

	if len(lines) != 10 {
		t.Fatalf("10 lines of log expected")
	}

}
