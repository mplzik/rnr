package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	rnr "github.com/mplzik/rnr/pkg"
)

type CliTask struct {
	Name string
	// Variant 1 -- nested tasks
	Parallelism int

	// Variant 2 -- a leaf task
	Command  []string
	Children []CliTask
}

var taskJsonPath = flag.String("taskJson", "-", "Path to the task json")

func parseTask(task CliTask) rnr.TaskInterface {
	// Nested task
	conversions := 0
	var ret rnr.TaskInterface

	if len(task.Children) > 0 {
		parallelism := task.Parallelism
		if parallelism == 0 {
			parallelism = 1
		}
		nested := rnr.NewNestedTask(task.Name, parallelism)
		// Recursively process the children
		for _, child := range task.Children {
			nested.Add(parseTask(child))
		}
		ret = nested
		conversions++
	}

	if len(task.Command) > 0 {
		cmd := rnr.NewShellTask(task.Name, task.Command...)
		ret = cmd
		conversions++
	}

	if conversions == 0 {
		log.Fatalf("Couldn't load task \"%s\".", task.Name)
	}

	if conversions > 1 {
		log.Fatalf("Task \"%s\" is of ambiguous type.", task.Name)
	}

	return ret
}

func main() {
	var cliTasks CliTask

	flag.Parse()

	fmt.Println("Hello, world!")

	taskJsonFile, err := os.Open(*taskJsonPath)
	if err != nil {
		log.Fatalf("Failed to open task file: %s", err.Error())
	}
	defer taskJsonFile.Close()

	taskJsonBytes, err := ioutil.ReadAll(taskJsonFile)
	if err != nil {
		log.Fatalf("Failed to unmarshal tasks file: %s", err.Error())
	}
	json.Unmarshal(taskJsonBytes, &cliTasks)

	fmt.Println(cliTasks)

	// Construct the tasks
	root := parseTask(cliTasks)

	job := rnr.NewJob(root)
	job.Start()

	go func() {
		for range time.Tick(time.Second) {
			fmt.Println("poll")
			job.Poll()
		}
	}()

	rnr.NewRnrWebserver(job)
	rnr := rnr.NewRnrWebserver(job)
	rnr.Start(":8080")

}
