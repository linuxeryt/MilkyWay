package main

import (
	"context"
	"milkyway/job"
	"sync"
)


func main() {
	var wg sync.WaitGroup

	taskCtx, _ := context.WithCancel(context.Background())
	wg.Add(1)
	go job.StartJobProcess(taskCtx, "devops") // start job process


	wg.Wait()
}