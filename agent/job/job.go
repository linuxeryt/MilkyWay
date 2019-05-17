package job

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"log"
	"milkyway/agent/job/register"
	"time"

	_ "milkyway/agent/job/module/all"		// 导入all包，会自动调用所有job模块init()，实现job注册
)

var (
	jobPrefix = "/milkyway/agent/job/"
)

type JobInfo struct {
	ModuleName	string					`json:"moduleName"`
	JobID		string					`json:"jobID"`
	Param 		map[string]interface{}	`json:"param"`
}


func StartJobProcess(ctx context.Context, agentID string) {
	log.Println("Start Job Process")

	jobKey := jobPrefix + agentID
	jobChan := startWatchJobKey(jobKey)
	resultChan := watchResultChan()

	loop(jobChan, resultChan, ctx)
}


func watchResultChan() chan map[string]interface{} {
	resultChan := make(chan map[string]interface{})

	go func() {
		select {
		case result := <-resultChan:
			log.Println(result)
		}
	}()

	return resultChan
}


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
				// log.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				jobChan <- ev.Kv.Value
			}
		}
	}()

	return jobChan
}


func loop(jobChan chan []byte, resultChan chan map[string]interface{},ctx context.Context) {
	// loop fetch job from jobChan
	for {
		select {
		case _job := <-jobChan: // {"moduleName": string, "param": map[string]{}interface}
			log.Printf("Received job: %s\n", string(_job))
			go callJobModule(_job, resultChan)
		case <-ctx.Done():
			log.Println("Job Process received exit signal.")
			log.Println("Job Process Exit.")
			return
		}
	}
}


func callJobModule(_job  []byte, resultChan chan map[string]interface{}) {
	var jobInfo JobInfo
	err := json.Unmarshal(_job, &jobInfo)
	if err != nil {
		log.Printf("Job Info Unmarshal error: %s\n",err)
		return
	}

	moduleName := jobInfo.ModuleName
	param := jobInfo.Param
	jobId := jobInfo.JobID

	log.Printf("param type: %T\n", param)

	jobModule, ok := register.ModuleMapOfJob[moduleName]
	if ok {
		jobModule.Run(jobId, param, resultChan)
	} else {
		log.Printf("Job module <%s> not found.\n", moduleName)
	}
}
