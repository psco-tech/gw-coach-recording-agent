package uploader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/psco-tech/gw-coach-recording-agent/models"
	"github.com/spf13/viper"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSixDigitGUID() string {
	rand.Seed(time.Now().UnixNano())

	guid := make([]byte, 6)
	for i := range guid {
		guid[i] = charset[rand.Intn(len(charset))]
	}
	return string(guid)
}

var (
	fileChan   chan models.UploadRecord
	once       sync.Once
	appConnect AppConnect
)

func Start() {
	once.Do(func() {
		fileChan = make(chan models.UploadRecord)
		go uploader(fileChan)
	})
}

func GetUploadRecordChannel() chan models.UploadRecord {
	return fileChan
}

func uploader(uploadRecords chan models.UploadRecord) {
	for ur := range uploadRecords {

		database, err := models.NewDatabase()
		appConfig, err := database.GetAppConfig()
		if err != nil || appConfig.AgentToken == "" {
			log.Printf("Could not get app Config: %s", err.Error())
			continue
		}

		appConnect := NewAppConnect(appConfig.AgentToken, viper.GetString("app_connect_host"))

		filename := filepath.Base(ur.FilePath)

		ur.Status = models.UploadStatusStarting
		database.Save(ur)
		log.Printf("Starting upload for record %d", ur.ID)

		var tempUploadRequest = *new(TempUploadUrlRequest)
		// Make sure filename is sufficiently unique
		tempUploadRequest.Filename = generateSixDigitGUID() + "_" + filename
		tempUploadRequest.ContentType = ur.ContentType

		tempUploadResponse, err := appConnect.GetTempUpload(tempUploadRequest)
		if err != nil || tempUploadResponse.URL == "" {
			log.Printf("Could not get temp upload URL")
			continue
		}

		log.Printf("Temp Upload URL is: %s", tempUploadResponse.URL)
		ur.Status = models.UploadStatusUploading
		database.Save(ur)
		log.Printf("Starting to upload data for for record %d", ur.ID)

		uploadFile, err := os.Open(ur.FilePath)
		if err != nil {
			log.Printf("Could not open file for upload: %s", err.Error())
			continue
		}
		defer uploadFile.Close()
		info, err := uploadFile.Stat()

		fileContent, err := ioutil.ReadFile(ur.FilePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			continue
		}

		// Create an HTTP PUT request with the file content
		req, err := http.NewRequest(http.MethodPut, tempUploadResponse.URL, bytes.NewReader(fileContent))
		req.Header.Set("Content-Type", ur.ContentType)
		req.Header.Set("Content-Length", fmt.Sprintf("%d", info.Size()))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error uploading file: %v", err)
			continue
		}
		defer resp.Body.Close()

		ur.Status = models.UploadStatusUploadTransferred
		database.Save(ur)
		log.Printf("Data all transferred for record %d", ur.ID)

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Printf("Error uploading file, status: %d, response: %s", resp.StatusCode, string(body))
			continue
		}

		switch ur.Type {
		case models.UploadRecordTypeCFS_AUDIO:
			var cfsAudio = *new(models.CFSAudio)
			cfsAudio.CallId = fmt.Sprintf("%d", ur.ID)
			cfsAudio, err := appConnect.FinalizeCFSUpload(tempUploadResponse.ObjectKey, cfsAudio)
			if err != nil {
				log.Printf("Error finalize file: ERROR: %s", err.Error())
				continue
			}
			break
		case models.UploadRecordTypeCAD:
			break
		}

		ur.Status = models.UploadStatusUploadFinalized
		database.Save(ur)
		log.Printf("Upload complete for record %d", ur.ID)

		// Remove the file after successful upload
		err = os.Remove(ur.FilePath)
		if err != nil {
			fmt.Printf("Error removing file: %v\n", err)
		}
	}
}
