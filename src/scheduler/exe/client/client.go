package main

import (
	"flag"
	"log"
	"scheduler/client"
)

func main() {

	log.SetFlags(log.Lmicroseconds)

	var command string
	const (
		usageCommand = "Command to run"
	)
	flag.StringVar(&command, "command", "", usageCommand)
	flag.StringVar(&command, "c", "", usageCommand+" (shorthand)")

	var list string
	const (
		usageList = "Retrieve a given task (by uuid) or all tasks"
	)
	flag.StringVar(&list, "list", "", usageList)
	flag.StringVar(&list, "l", "", usageList+" (shorthand)")

	var url string
	const (
		defaultUrl = "http://localhost:8080"
		usageUrl   = "URL of server"
	)
	flag.StringVar(&url, "url", defaultUrl, usageUrl)
	flag.StringVar(&url, "u", defaultUrl, usageUrl+" (shorthand)")

	flag.Parse()

	client := client.Client{Url: url}

	if command != "" {
		task, err := client.Execute(command)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(task.Uuid)
	} else if list != "" {
		task, err := client.GetTask(list)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(task)
	} else {
		tasks, err := client.GetTasks()
		if err != nil {
			log.Fatal(err)
		}

		for _, task := range tasks {
			log.Println(task)
		}
	}

}
