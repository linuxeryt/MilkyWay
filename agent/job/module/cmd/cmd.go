package cmd

import (
	"bytes"
	"encoding/json"
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
		log.Println("error: ", err)
		result = errstd.String()
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

	log.Printf("Call cmd module, params: %s.\n", param)

	path := param["path"].(string)
	_args := param["args"].([]interface{})

	args := make([]string, len(_args))
	for index, value := range _args {
		args[index] = fmt.Sprint(value)
	}

	cmd.runCmd(path, args)
	cmd.Return(resultChan)
}

func (cmd *Cmd) Return(resultChan chan map[string]interface{}) {

	b, err := json.Marshal(cmd)
	if err != nil {
		log.Println("err: ", err.Error())
	} else {
		log.Println(b)
	}

	res := map[string]interface{}{
		"JobID": cmd.JobID,
		"Status": cmd.Status,
		"Result": cmd.Result,
	}

	resultChan <- res
}

func init() {
	register.Register("cmd", &Cmd{})
}