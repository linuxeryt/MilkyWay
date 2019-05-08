package cmd

import (
	"fmt"
	"milkyway/agent/job/register"
)

type Cmd struct{
	param interface{}
}


func (cmd Cmd) Run(param interface{}) {
	fmt.Println("run cmd")
}


func (cmd Cmd) Return() {

}

func init() {
	register.Register("cmd", Cmd{})
}