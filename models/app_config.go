package models

type AppConfig struct {
	// Key from the Coach Server that allows for upload
	AgentKey string

	UploadDir string
}
