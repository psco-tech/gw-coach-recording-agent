package models

import "gorm.io/gorm"

// An AESRecordingDevice holds the configuration data for one
// virtual station that is used internally to monitor the conversation
type AESRecordingDevice struct {
	gorm.Model

	// The extension to use as the recording device
	Extension string

	// The security code of the virtual station
	Password string
}
