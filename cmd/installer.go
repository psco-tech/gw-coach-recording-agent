package cmd

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const serviceName = "GovWorx Data Agent"
const serviceDescription = "Service to upload configured data to CommsCoach and BluAssist"

var logger service.Logger

// See example here
// https://github.com/kardianos/service/blob/master/example/runner/runner.go

type program struct{}

func (p program) Start(s service.Service) error {
	fmt.Println("Starting the service")
	go p.run()
	return nil
}

func logsDir() (string, error) {

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
		return "", err
	}
	baseDir := filepath.Dir(executablePath)
	outputDir := filepath.Join(baseDir, "logs")

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Error creating 'uploads' directory: %v", err)
		return "", err
	}

	return outputDir, nil
}

func (p *program) run() {

	executablePath, err := os.Executable()
	fmt.Println("Executable path: " + executablePath)
	c := *exec.Command(executablePath)
	if err != nil {
		fmt.Println("Error getting executable path: %v", err.Error())
	}

	fmt.Println("Executable path: " + c.Path)

	logdir, err := logsDir()
	if err != nil {
		fmt.Println("Error opening logdir: %v", err.Error())
	}
	fmt.Println("Logdir: " + logdir)
	f, err := os.OpenFile(filepath.Join(logdir, "gw-error_log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		logger.Warningf("Failed to open std err log: %v", err)
		return
	}
	defer f.Close()
	c.Stderr = f

	outf, err := os.OpenFile(filepath.Join(logdir, "gw-output_log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		logger.Warningf("Failed to open std out output log: %v", err)
		return
	}
	defer f.Close()
	c.Stdout = outf

	err = c.Run()
	if err != nil {
		fmt.Println("Error running: %v", err.Error())
	}

	fmt.Println("Service started")
}

func (p program) Stop(s service.Service) error {
	return nil
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Windows service",
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceConfig := &service.Config{
			Name:        serviceName,
			DisplayName: serviceName,
			Description: serviceDescription,
		}

		fmt.Println("Registering the service")

		prg := &program{}
		s, err := service.New(prg, serviceConfig)
		if err != nil {
			fmt.Println("Cannot create the service: " + err.Error())
		}
		err = s.Run()
		if err != nil {
			fmt.Println("Cannot start the service: " + err.Error())
		}
		return err
	},
}
