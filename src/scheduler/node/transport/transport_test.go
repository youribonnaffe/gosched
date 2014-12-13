package transport

import (
	"net/http"
	"net/http/httptest"
	"scheduler/server"
	"scheduler/shared"
	"testing"
)

func TestListTasks(t *testing.T) {
	ts, store := startHttpServer()
	store.Tasks["abc"] = &server.LockedTask{Task: &shared.Task{Executable: "ls", Status: shared.Pending}}
	store.Tasks["bde"] = &server.LockedTask{Task: &shared.Task{Executable: "ls", Status: shared.Finished}}

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
	ts, store := startHttpServer()
	store.Tasks["abc"] = &server.LockedTask{Task: &shared.Task{Status: shared.Pending}}

	transport := HttpNodeTransport{ts.URL}
	err := transport.Update(shared.Task{Uuid: "abc", Status: shared.Finished})

	if err != nil {
		t.Fatal(err)
	}

	if store.Tasks["abc"].Task.Status != shared.Finished {
		t.Fatal("Finished status expected")
	}

	ts.Close()
}

func startHttpServer() (*httptest.Server, *server.Store) {
	store := server.NewStore()

	ts := httptest.NewServer(http.HandlerFunc(store.TaskHandler))

	return ts, store
}
