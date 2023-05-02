package models

import "gorm.io/gorm"

type UploadStatus string

const (
	UploadStatusQueued   UploadStatus = "QUEUED"
	UploadStatusUploading   UploadStatus = "UPLOADING"
	UploadStatusUploadTransferred UploadStatus = "UPLOAD_TRANSFERRED"
	UploadStatusUploadFinalized  UploadStatus = "UPLOAD_FINALIZED"
)

type UploadRecord struct {
	gorm.Model
	// Key from the Coach Server that allows for upload
	FilePath string

	Status UploadStatus

	ContentType string
}
