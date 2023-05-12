package models

import "gorm.io/gorm"

type PassiveMonitoringConfig struct {
	gorm.Model

	MTU int

	NetworkInterfaceName string
}
