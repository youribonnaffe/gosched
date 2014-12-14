package node

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"scheduler/node/transport"
	"scheduler/shared"
	"time"
)

type Node struct {
	Size  int
	tasks chan shared.Task
	t     transport.Protocol
}

func Start(size int, t transport.Protocol) Node {
	tasks := make(chan shared.Task)
	for i := 0; i < size; i++ {
		go worker(i, t, tasks)
	}
	return Node{Size: size, t: t, tasks: tasks}
}

func (node Node) Run() bool {
	tasks, err := node.t.ListTasks(shared.Pending)
	if err == nil {
		if len(tasks) == 0 {
			return false
		}
		log.Println("Got", len(tasks), "tasks to run")
		for _, task := range tasks {
			node.tasks <- task
		}
		return true
	} else {
		log.Println("Could not contact server", err)
		return false
	}

}

func worker(id int, t transport.Protocol, tasks chan shared.Task) {
	for task := range tasks {
		log.Println("working on", id, &task, task.Uuid)

		err := t.Update(shared.Task{Status: shared.Running, Uuid: task.Uuid})

		if err != nil {
			log.Println("Cannot start task, probably already running somewhere", task.Uuid)
			break
		}

		startTime := time.Now()
		cmd := exec.Command(task.Executable)
		stdout, _ := cmd.StdoutPipe()
		if err := cmd.Start(); err != nil {
			log.Println(err)
		}
		var out bytes.Buffer
		waitCmdOutput := make(chan bool)
		go func() {
			bufCmdOut := bufio.NewReader(stdout)
			for {
				line, _, err := bufCmdOut.ReadLine()
				if err != nil {
					waitCmdOutput <- true
					break
				} else {
					log.Println("###", string(line))
					out.Write(line)
					// log streaming here? TODO missing \n
				}
			}
		}()
		<-waitCmdOutput
		var executionDuration time.Duration
		if err := cmd.Wait(); err != nil {
			log.Println(err)
		} else {
			executionDuration = time.Since(startTime)
			log.Println("done working on", id, task.Uuid, "output is", out.String())
		}

		t.Update(shared.Task{Uuid: task.Uuid, Status: shared.Finished,
			Output:    out.String(),
			StartTime: startTime, ExecutionDuration: executionDuration})
	}
}
