package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/viper"
)

// WaitGroup for background tasksWg
var tasksWg sync.WaitGroup

func main() {
	applicationContext, shutdown := context.WithCancel(context.Background())

	// Gracefully shut down when returning
	defer func() {
		shutdown()
		tasksWg.Wait()
	}()

	// Read configuration file, if any and apply some defaults
	err := readInConfiguration()
	if err != nil {
		log.Printf("Failed to read configuration: %s\n", err)
		return
	}

	// Get RTP listening up and running
	rtpReceiverTask := NewRTPReceiverTask(applicationContext, &tasksWg)
	err = rtpReceiverTask.Start(viper.GetUint("rtp_receiver_count"))
	if err != nil {
		log.Printf("Failed to start RTP receiver Task: %s\n", err)
		return
	}

	// Connect to the PBX
	pbxConnectionTask := NewPBXConnectionTask(applicationContext, &tasksWg)
	pbxConnectionTask.SetRecorderPool(rtpReceiverTask)

	err = pbxConnectionTask.Start(viper.GetString("pbx_type"))
	if err != nil {
		log.Printf("Failed to start PBX connection Task: %s\n", err)
		return
	}

	// Setup application exit conditions
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for any of the signals to terminate the program
	<-sigs

	log.Printf("Application shutdown requested\n")
}

func readInConfiguration() error {
	viper.SetConfigName("call_recording_agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Global defaults
	// application_id is used in the StartApplicationSession CSTA/ACSE message
	viper.SetDefault("application_id", "CRA")

	err := viper.ReadInConfig()

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Printf("No config file found, using defaults\n")
	} else {
		return err
	}

	return nil
}
