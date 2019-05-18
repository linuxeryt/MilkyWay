package cmd

import (
	"bytes"
	"fmt"
	"log"
	"milkyway/agent/job/register"
	"os/exec"
)

type Cmd struct{
	JobID	string
	Status	bool
	Result	string
}

func (cmd *Cmd) runCmd(path string, args []string) {
	var out bytes.Buffer
	var errstd bytes.Buffer
	var result string
	var status bool

	command := exec.Command(path, args...)
	command.Stdout = &out
	command.Stderr = &errstd

	err := command.Run()
	if err != nil {
		result = err.Error()
		status = false
	} else {
		result = out.String()
		status = true
	}

	cmd.Result = result
	cmd.Status = status
}


// param: {"path": string, "args": []interface{}}
func (cmd *Cmd) Run(jobID string, param map[string]interface{}, resultChan chan map[string]interface{}) {

	cmd.JobID = jobID

	log.Printf("[job]  JobID: %s; cmd module running.\n", jobID)

	path := param["path"].(string)
	_args := param["args"].([]interface{})

	// 将[]interface{}中的转储入[]string中
	args := make([]string, len(_args))
	for index, value := range _args {
		args[index] = fmt.Sprint(value)
	}

	cmd.runCmd(path, args)
}

func (cmd *Cmd) Return() map[string]interface{} {
	res := map[string]interface{}{
		"JobID": cmd.JobID,
		"Status": cmd.Status,
		"Result": cmd.Result,
	}

	log.Printf("[Job]  JobID: %s; Result: %s\n", cmd.JobID, cmd.Result)
	return res
}

func init() {
	register.Register("cmd", &Cmd{})
}