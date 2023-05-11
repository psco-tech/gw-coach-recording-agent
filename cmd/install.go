package cmd

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "GW_Coach_Agent"

var cachedServiceManager *mgr.Mgr

func serviceManager() (*mgr.Mgr, error) {
	if cachedServiceManager != nil {
		return cachedServiceManager, nil
	}
	m, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	cachedServiceManager = m
	return cachedServiceManager, nil
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Windows service",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := serviceManager()
		if err != nil {
			return err
		}

		executablePath, err := os.Executable()
		if err != nil {
			return err
		}

		service, err := m.OpenService(serviceName)
		if err == nil {
			status, err := service.Query()
			if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
				service.Close()
				return err
			}
			if status.State != svc.Stopped && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
				service.Close()
				return errors.New("service already installed and running")
			}
			err = service.Delete()
			service.Close()
			if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
				return err
			}
			for {
				service, err = m.OpenService(serviceName)
				if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
					break
				}
				service.Close()
				time.Sleep(time.Second / 3)
			}
		}

		config := mgr.Config{
			ServiceType:  windows.SERVICE_WIN32_OWN_PROCESS,
			StartType:    mgr.StartAutomatic,
			ErrorControl: mgr.ErrorNormal,
			Dependencies: []string{"Nsi", "TcpIp"},
			Description:  "GovWorx Coach Recording Agent",
			DisplayName:  "GovWorx Coach Recording Agent",
			SidType:      windows.SERVICE_SID_TYPE_UNRESTRICTED,
		}
		service, err = m.CreateService(serviceName, executablePath, config)
		if err != nil {
			return err
		}

		err = service.Start()
		if err != nil {
			log.Printf("Failed to install service: %s\n", err)
			return err
		}
		service.Close()

		return err
	},
}
