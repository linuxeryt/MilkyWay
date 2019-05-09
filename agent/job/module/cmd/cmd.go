package cmd

import (
	"bytes"
	"fmt"
	"log"
	"milkyway/agent/job/register"
	"os/exec"
)

type Cmd struct{
	param interface{}
}

func runCmd(path string, args []string) {
	var out bytes.Buffer

	command := exec.Command(path, args...)
	command.Stdout = &out

	err := command.Run()
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(out.String())
	}
}

// param: {"path": string, "args": []string}
func (cmd Cmd) Run(params interface{}) {
	var _param map[string]interface{}
	_param = params.(map[string]interface{})

	path := _param["path"].(string)
	args := _param["args"].([]string)
	runCmd(path, args)
}


func (cmd Cmd) Return() {

}

func init() {
	register.Register("cmd", Cmd{})
}