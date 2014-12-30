package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"scheduler/client"
	"scheduler/shared"
	"sort"
)

func main() {

	command, list, url, output := parseArguments()

	client := client.Client{Url: url}

	if command != "" {
		task, err := client.Execute(command)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not execute command %s\n", err)
		}

		fmt.Println(task.Uuid)

	} else if list != "" {
		task, err := client.GetTask(list)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(task.String())

	} else if output != "" {
		fromLine := 0

		for {
			output, err := client.Tail(output, fromLine)
			if err != nil {
				log.Fatal(err)
			}

			for _, line := range output {
				fmt.Println(line)
			}

			fromLine += len(output)
		}

	} else {
		tasks, err := client.GetTasks()
		if err != nil {
			log.Fatal(err)
		}

		sort.Sort(shared.BySubmittedTime(tasks))
		for _, task := range tasks {
			fmt.Println(task.String())
		}
	}

}

func parseArguments() (command string, list string, url string, output string) {
	const (
		usageCommand = "Command to run"
	)
	flag.StringVar(&command, "command", "", usageCommand)
	flag.StringVar(&command, "c", "", usageCommand+" (shorthand)")

	const (
		usageList = "Retrieve a given task (by uuid) or all tasks"
	)
	flag.StringVar(&list, "list", "", usageList)
	flag.StringVar(&list, "l", "", usageList+" (shorthand)")

	const (
		usageOutput = "Get output of a task"
	)
	flag.StringVar(&output, "output", "", usageOutput)
	flag.StringVar(&output, "o", "", usageOutput+" (shorthand)")

	const (
		defaultUrl = "http://localhost:8080"
		usageUrl   = "URL of server"
	)
	flag.StringVar(&url, "url", defaultUrl, usageUrl)
	flag.StringVar(&url, "u", defaultUrl, usageUrl+" (shorthand)")

	flag.Parse()

	return command, list, url, output
}
