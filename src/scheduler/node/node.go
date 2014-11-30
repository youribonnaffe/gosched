package node

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"scheduler/node/transport"
	"scheduler/shared"
)

type Node struct {
	Size  int
	tasks chan *shared.Task
	t     transport.Protocol
}

func Start(size int, t transport.Protocol) Node {
	tasks := make(chan *shared.Task)
	for i := 0; i < size; i++ {
		go worker(t, tasks)
	}
	return Node{Size: size, t: t, tasks: tasks}
}

func (node Node) Run() bool {
	tasks, err := node.t.ListTasks(shared.Pending)
	if err == nil {
		ret := false
		for id, task := range tasks {
			log.Println("starting on", id)
			node.tasks <- &task
			log.Println("started on", id)
			ret = true
		}
		return ret
	} else {
		log.Println("Could not contact server", err)
		return false
	}

}

func worker(t transport.Protocol, tasks <-chan *shared.Task) {
	for task := range tasks {
		log.Println("working on", task.Uuid)

		t.Update(shared.Task{Status: shared.Running, Uuid: task.Uuid})

		// TODO check it is ok to start working on it

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
				}
			}
		}()
		<-waitCmdOutput
		if err := cmd.Wait(); err != nil {
			log.Println(err)
		} else {
			log.Println("done working on", task.Uuid, "output is", out.String())
		}

		t.Update(shared.Task{Uuid: task.Uuid, Status: shared.Finished, Output: out.String()})
	}
}
