package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/judwhite/go-svc"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
	"github.com/psco-tech/gw-coach-recording-agent/pbx/avaya"
	"github.com/psco-tech/gw-coach-recording-agent/pbx/osbiz"
	"github.com/psco-tech/gw-coach-recording-agent/rtp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(readInConfiguration)
}

var rootCmd = &cobra.Command{
	Use:   "cra",
	Short: "CRA is a call recording agent",
	Run: func(cmd *cobra.Command, args []string) {
		craService := &callRecordingAgentService{}

		if err := svc.Run(craService); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type callRecordingAgentService struct {
	wg   sync.WaitGroup
	ctx  context.Context
	quit context.CancelFunc

	recorderPool rtp.RecorderPool
	pbx          pbx.PBX
}

func (c *callRecordingAgentService) Init(env svc.Environment) error {
	c.ctx, c.quit = context.WithCancel(context.Background())

	err := c.registerPBXImplementations()
	if err != nil {
		return fmt.Errorf("failed to register PBX implementations: %w", err)
	}

	// Create the recorder pool
	c.recorderPool, err = rtp.NewRecorderPool(viper.GetUint("rtp.recorder_count"), c.ctx)
	if err != nil {
		return fmt.Errorf("failed to create the recorder pool: %w", err)
	}

	// Instantiate the PBX connection
	c.pbx, err = pbx.New(viper.GetString("pbx_type"), c.ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate PBX implementation: %w", err)
	}

	return nil
}

func (c *callRecordingAgentService) Start() error {
	// Start the RecorderPool first so the PBX will never request from it without it running
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := c.recorderPool.Start()
		if err != nil {
			log.Printf("Recorder pool error: %s\n", err)
		}
	}()

	// Connect to the PBX
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := c.reestablishPBXConnection()
		if err != nil {
			log.Printf("Recorder pool error: %s\n", err)
		}
	}()

	return nil
}

func (c *callRecordingAgentService) Stop() error {
	log.Printf("Application shutdown requested\n")
	c.quit()

	// Wait for background activity to finish
	c.wg.Wait()
	return nil
}

func readInConfiguration() {
	viper.SetConfigName("call_recording_agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Global defaults
	// application_id is used in the StartApplicationSession CSTA/ACSE message
	viper.SetDefault("application_id", "CRA")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("No config file found, using defaults\n")
		} else {
			log.Printf("Error reading configuration file: %s\n", err)
			os.Exit(1)
		}
	}
}

func (c *callRecordingAgentService) registerPBXImplementations() error {
	pbx.RegisterImplementation("osbiz", &osbiz.OSBiz{})
	pbx.RegisterImplementation("avaya_aes", &avaya.AvayaAES{})

	return nil
}

func (c *callRecordingAgentService) reestablishPBXConnection() error {
	const connectionRetryTimeout = 30 * time.Second
	log.Printf("Starting PBX connection handler\n")

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Closing PBX Session\n")
			c.pbx.Close()
			return nil
		default:
			log.Printf("Trying to connect to PBX...\n")
			_, err := c.pbx.Connect()

			if err != nil {
				log.Printf("Failed to connect to PBX, retry in %ds: %s\n", connectionRetryTimeout/time.Second, err)

				select {
				case <-c.ctx.Done():
					return nil
				// Try to connect again after connectionRetryTimeout
				case <-time.After(connectionRetryTimeout):
					log.Printf("Retrying PBX connection\n")
					continue
				}
			}

			log.Printf("Successfully connected to PBX\n")
			err = c.pbx.Serve(c.recorderPool)

			// When Serve() returns an err the connection is lost or closed
			if err != nil {
				log.Printf("PBX connection closed, reconnect in %ds: %s\n", connectionRetryTimeout/time.Second, err)

				select {
				case <-c.ctx.Done():
					return nil
				// Try to connect again after connectionRetryTimeout
				case <-time.After(connectionRetryTimeout):
					log.Printf("Retrying PBX connection\n")
					continue
				}
			}

			// If Serve() closed without an error we're done
			return nil
		}
	}
}
