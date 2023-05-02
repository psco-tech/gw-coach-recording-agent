package models

import "gorm.io/gorm"

type AppConfig struct {
	gorm.Model
	// Key from the Coach Server that allows for upload
	AgentToken string

	UploadDir string
}
