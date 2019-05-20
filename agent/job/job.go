package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"milkyway/agent/config"
	"milkyway/agent/job/register"
	"time"

	_ "milkyway/agent/job/module/all" // 导入all包，会自动调用所有job模块init()，实现job注册
)

var (
	jobPrefix = "/milkyway/agent/job/"
	agentConfig = new(config.GlobalConfig)
)

type JobInfo struct {
	ModuleName string                 `json:"moduleName"`
	JobID      string                 `json:"jobID"`
	Param      map[string]interface{} `json:"param"`
}

func StartJobProcess(ctx context.Context, agentID string, config *config.GlobalConfig) {
	log.Println("[job]  Start Job Process.")
	log.Printf("[job] Job received AgentConfig: %s\n", *config)

	agentConfig = config
	jobKey := jobPrefix + agentID

	// jobChan
	// loop从该通道接收任务
	// 监听jobKey的gouroutine发送任务至该通道
	jobChan := startWatchJobProcess(jobKey, ctx)

	// resultChan
	// job模块的运行结果将发送到该通道
	// 监听该通道的goroutine将结果发送至etcd集群
	resultChan := startSendResultProcess(ctx)

	loop(jobChan, resultChan, ctx)
}


// 启动 SendResult 进程
// 监听 resultChan 并 将结果发送至etcd集群
func startSendResultProcess(ctx context.Context) chan map[string]interface{} {
	resultChan := make(chan map[string]interface{})

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   agentConfig.EtcdEndPoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Println(err)
	}

	go func() {
		for {
			select {
			case result := <-resultChan:
				jobID := result["JobID"].(string)

				// 根据jobID构造结果返回的key
				key := "/milkyway/server/job/" + jobID
				v, _ := json.Marshal(result)
				cli.Put(ctx, key, string(v))
			case <-ctx.Done():
				log.Println("[job]  SendResult process received exit signal.")
				log.Println("[job]  SendResult process exit.")
				return
			}
		}

	}()
	return resultChan
}

// start WatchJob process to watch job key
// 启动 WatchJob 子进程
// 用于从etcd集群接收job任务并将任务发送至 jobChan
func startWatchJobProcess(jobKey string, ctx context.Context) chan []byte {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:    agentConfig.EtcdEndPoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Println(err)
	}

	rch := cli.Watch(ctx, jobKey)
	jobChan := make(chan []byte)

	go func() {
		defer cli.Close()

		log.Printf("[job]  WatchJob process watching job key %s\n", jobKey)
		for {
			select {
			case wresp := <-rch:
				for _, ev := range wresp.Events {
					log.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					jobChan <- ev.Kv.Value
				}
			case <-ctx.Done():
				log.Println("[job]  WatchJob process received exit signal.")
				log.Println("[job]  WatchJob process exit.")
				return
			}
		}
	}()

	return jobChan
}


func loop(jobChan chan []byte, resultChan chan map[string]interface{}, ctx context.Context) {

	// loop fetch job from jobChan
	for {
		select {
		case _job := <-jobChan: // {"moduleName": string, "jobID": string, "param": map[string]{}interface}
			log.Printf("[Job]  Received job with data: %s\n", string(_job))
			go callJobModule(_job, resultChan)
		case <-ctx.Done():
			log.Println("[Job]  job main process received exit signal.")
			log.Println("[Job]  job main process exit.")
			return
		}
	}
}


func callJobModule(_job []byte, resultChan chan map[string]interface{}) {
	var jobInfo JobInfo
	err := json.Unmarshal(_job, &jobInfo)
	if err != nil {
		log.Printf("[Job]  Job Info Unmarshal error: %s\n", err)
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
		fmt.Fprintf(_result, "[Job] JobID: %s; Job module '%s' not found.", jobId, moduleName)
		result := map[string]interface{}{
			"JobID":  jobId,
			"Status": false,
			"Result": _result.String(),
		}

		log.Printf("[Job] JobID: %s; Result: %s\n", jobId, _result)

		resultChan <- result

	}
}
