package main

import (
	"context"
	"milkyway/agent/job"
	"sync"
)


func main() {
	var wg sync.WaitGroup

	// start job process
	taskCtx, _ := context.WithCancel(context.Background())
	wg.Add(1)
	go job.StartJobProcess(taskCtx, "devops")

	// start data collect process
	// wg.Add(1)
	// go

	// start process of watch global config
	//wg.Add(1)
	//go

	wg.Wait()
}