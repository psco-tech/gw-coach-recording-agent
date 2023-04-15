package models

import "gorm.io/gorm"

type AESRecordingDevice struct {
	gorm.Model

	// The extension to use as the recording device
	Extension string

	// The security code of the virtual station
	Password string
}
