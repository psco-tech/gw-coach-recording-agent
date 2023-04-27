package models

import "gorm.io/gorm"

// An AESRecordingDevice holds the configuration data for one
// virtual station that is used internally to monitor the conversation
type PBXConnectionCredentials struct {
	gorm.Model

	// The vendor the PBX
	PbxType string

	// The host of the PBX
	Host string

	// The port to connect to
	Port string

	// The username to connect
	Username string

	// The password to connect
	Password string
}
