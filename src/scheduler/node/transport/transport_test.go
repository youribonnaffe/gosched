package transport

import (
	"net/http"
	"net/http/httptest"
	"scheduler/server"
	"scheduler/shared"
	"testing"
)

func TestListTasks(t *testing.T) {
	ts, scheduler := startHttpServer()
	scheduler.CreateTask(shared.Task{Executable: "ls"})
	finishedTask := scheduler.CreateTask(shared.Task{Executable: "ls"})

	finishedTask.Status = shared.Finished
	scheduler.UpdateTask(finishedTask)

	transport := HttpNodeTransport{ts.URL}
	tasks, err := transport.ListTasks(shared.Pending)

	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatal("One task expected")
	}

	if tasks[0].Executable != "ls" {
		t.Fatal("Executable in task expected")
	}

	ts.Close()
}

// polling now
//func TestEmptyListTasks(t *testing.T) {
//	ts, _ := startHttpServer()
//
//	transport := HttpNodeTransport{ts.URL}
//	emptyTasks, err := transport.ListTasks(shared.Pending)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if len(emptyTasks) != 0 {
//		t.Fatal("No task expected")
//	}
//
//	ts.Close()
//}

func TestUpdateTask(t *testing.T) {
	ts, scheduler := startHttpServer()
	task := scheduler.CreateTask(shared.Task{Executable: "ls"})

	transport := HttpNodeTransport{ts.URL}
	err := transport.Update(shared.Task{Uuid: task.Uuid, Status: shared.Finished})

	if err != nil {
		t.Fatal(err)
	}

	task, _ = scheduler.GetTask(task.Uuid)
	if task.Status != shared.Finished {
		t.Fatal("Finished status expected")
	}

	ts.Close()
}

func startHttpServer() (*httptest.Server, *server.Scheduler) {
	scheduler := server.NewScheduler()

	ts := httptest.NewServer(http.HandlerFunc(scheduler.TaskHandler))

	return ts, scheduler
}
