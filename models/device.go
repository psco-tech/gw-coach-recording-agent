package models

import (
	"time"

	"gorm.io/gorm"
)

type Device struct {
	gorm.Model

	// The human readable extension of this device
	Extension string

	// A Description for management purposes
	Description string

	// The internal device ID this device can be accessed through on the PBX
	DeviceID string

	// Should this device be recorded?
	RecordCalls bool

	// Last known CSTA cross reference ID
	CrossReferenceID string

	// The time this device has had a call recorded from for the last time
	LastRecordedCall time.Time
}
