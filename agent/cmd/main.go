package main

import (
	"context"
	"flag"
	"log"
	"milkyway/agent/config"
	"milkyway/agent/job"
	"os"
	"strings"
	"sync"
)

var (
	etcdEndPoints = flag.String("etcdEndpoints", "", "etcd endpoints")
	agentID       = flag.String("hostid", "", "agent唯一标识，使用外网IP")
)

var (
	AgentConfig = config.AgentConfig
)

func init() {
	flag.Parse()

	// 使用hostname作为HostID
	if *agentID == "" {
		var err error
		*agentID, err = os.Hostname()

		if err != nil {
			log.Fatal(err)
		} else {
			AgentConfig.AgentID = *agentID
		}
	}

	if *etcdEndPoints == "" {
		flag.Usage()
		os.Exit(1)
	} else {
		temp := strings.Split(*etcdEndPoints, ",")
		AgentConfig.EtcdEndPoints = temp
	}

	log.Printf("Init AgentID: %s\n", *agentID)
	log.Printf("Init EtcdEndPoints: %s\n", *etcdEndPoints)
	log.Printf("Init AgentConfig: %s\n", *AgentConfig)
}

func main() {
	var wg sync.WaitGroup

	// start job process
	taskCtx, _ := context.WithCancel(context.Background())
	wg.Add(1)
	go job.StartJobProcess(taskCtx, "devops", AgentConfig)

	// start data collect process
	// wg.Add(1)
	// go

	// start process of watch global config
	//wg.Add(1)
	//go

	wg.Wait()
}
