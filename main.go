package main

import (
	"github.com/psco-tech/gw-coach-recording-agent/cmd"
	"github.com/psco-tech/gw-coach-recording-agent/configserver"
	"github.com/psco-tech/gw-coach-recording-agent/uploader"
)

func main() {
	configserver.Start()
	uploader.Start()
	cmd.Execute()
}
