package main

import (
	"github.com/psco-tech/gw-coach-recording-agent/cmd"
	"github.com/psco-tech/gw-coach-recording-agent/configserver"
)

func main() {
	configserver.Start()
	cmd.Execute()
}
