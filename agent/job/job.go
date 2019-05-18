package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	log.Println("[job]  Start Job Process.")

	jobKey := jobPrefix + agentID

	// jobChan
	// loop从该通道接收任务
	// 监听jobKey的gouroutine发送任务至该通道
	jobChan := startWatchJobKey(jobKey)

	// resultChan
	// job模块的运行结果将发送到该通道
	// 监听该通道的goroutine将结果发送至etcd集群
	resultChan := watchResultChan(ctx)

	loop(jobChan, resultChan, ctx)
}


func watchResultChan(ctx context.Context) chan map[string]interface{} {
	resultChan := make(chan map[string]interface{})

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Println(err)
	}

	go func() {
		select {
		case result := <-resultChan:
			jobID := result["JobID"].(string)

			// 根据jobID构造结果返回的key
			key := "/milkyway/server/job/" + jobID
			v, _ := json.Marshal(result)
			cli.Put(ctx, key, string(v))
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
		case _job := <-jobChan: // {"moduleName": string, "jobID": string, "param": map[string]{}interface}
			log.Printf("[Job]  Received job with data: %s\n", string(_job))
			go callJobModule(_job, resultChan)
		case <-ctx.Done():
			log.Println("[Job]  Job Process received exit signal.")
			log.Println("[Job]  Job Process Exit.")
			return
		}
	}
}


func callJobModule(_job  []byte, resultChan chan map[string]interface{}) {
	var jobInfo JobInfo
	err := json.Unmarshal(_job, &jobInfo)
	if err != nil {
		log.Printf("[Job]  Job Info Unmarshal error: %s\n",err)
		return
	}

	moduleName := jobInfo.ModuleName
	param := jobInfo.Param
	jobId := jobInfo.JobID

	jobModule, ok := register.ModuleMapOfJob[moduleName]
	if ok {
		jobModule.Run(jobId, param, resultChan)
		result := jobModule.Return()
		resultChan <- result
	} else {
		_result := new(bytes.Buffer)
		fmt.Fprintf(_result,"[Job] JobID: %s; Job module '%s' not found.", jobId, moduleName)
		result := map[string]interface{}{
			"JobID": jobId,
			"Status": false,
			"Result": _result.String(),
		}
		resultChan <- result

	}
}