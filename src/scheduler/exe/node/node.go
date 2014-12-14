package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"scheduler/node"
	"scheduler/node/transport"
	"time"
)

func main() {

	// list all jobs
	// take a pending one
	// run it
	// mark it as finished
	// and again

	runtime.GOMAXPROCS(8)

	log.SetFlags(log.Lmicroseconds)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Fatal("sig", sig)
		}
	}()

	transport := transport.HttpNodeTransport{Url: "http://localhost:8080"}
	node := node.Start(8, transport)
	log.Println("Starting node with", node.Size, "workers")
	for {
		if !node.Run() {
			time.Sleep(1 * time.Millisecond)
		} else {
			time.Sleep(100 * time.Millisecond) // to avoid too much concurrent startup of tasks
		}
	}

	log.Println("end")
}
