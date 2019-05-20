package main

import (
	"context"
	"flag"
	"log"
	"milkyway/agent/job"
	"sync"
)

var (
	EtcdEndPoints = flag.String("etcdEndpoints", "", "etcd endpoints")
	HostID        = flag.String("hostid", "", "agent唯一标识，使用外网IP")
)

func init() {
	flag.Parse()

	if *EtcdEndPoints == "" {
		log.Fatal("Need EtcdEndPoints param")
		flag.Usage()
	}

	if *HostID == "" {
		// 使用hostname 作为HostID
	}
}

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
