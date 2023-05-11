package models

import (
	"time"

	"gorm.io/gorm"
)

type UploadStatus string

type UploadRecordType string

const (
	UploadStatusQueued            UploadStatus = "QUEUED"
	UploadStatusStarting          UploadStatus = "UPLOAD_STARTING"
	UploadStatusUploading         UploadStatus = "UPLOADING"
	UploadStatusUploadTransferred UploadStatus = "UPLOAD_TRANSFERRED"
	UploadStatusUploadFinalized   UploadStatus = "UPLOAD_FINALIZED"
)

const (
	UploadRecordTypeCFS_AUDIO UploadRecordType = "CFS_AUDIO"
	UploadRecordTypeCAD       UploadRecordType = "CAD"
)

type UploadRecord struct {
	gorm.Model
	// Key from the Coach Server that allows for upload
	FilePath string

	Status UploadStatus

	ContentType string

	Details string
	Begin   time.Time
	End     time.Time

	Type UploadRecordType
}
