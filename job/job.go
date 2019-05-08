package job

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

var (
	jobPrefix = "/milkyway/agent/job/"
)

// start watch job key
func startWatchJobKey(jobKey string) chan []byte {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Println(err)
	}

	rch := cli.Watch(context.Background(), jobKey)
	jobChan := make(chan []byte)

	go func() {
		defer cli.Close()

		log.Printf("Job Process watching job key %s\n", jobKey)

		for wresp := range rch {
			for _, ev := range wresp.Events {
				log.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				jobChan <- ev.Kv.Value
			}
		}
	}()

	return jobChan
}

func StartJobProcess(ctx context.Context, agentID string) {
	log.Println("Start Job Process")

	jobKey := jobPrefix + agentID
	jobChan := startWatchJobKey(jobKey)

	// loop fetch job from jobChan
	for {
		select {
		case task := <-jobChan:
			log.Printf("Received job: %s\n", string(task))
		case <-ctx.Done():
			return
		}
	}
}