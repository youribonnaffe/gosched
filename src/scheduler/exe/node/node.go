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

	runtime.GOMAXPROCS(6)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Fatal("sig", sig)
		}
	}()

	transport := transport.HttpNodeTransport{Url: "http://localhost:8080"}
	node := node.Start(100, transport)
	log.Println("Starting node with", node.Size, "workers")
	for {
		if !node.Run() {
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Println("end")
}
