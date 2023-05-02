package models

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func init() {
	viper.SetDefault("config_path", "config.db")
}

type DB struct {
	gormDB *gorm.DB
}

var database *DB

func NewDatabase() (*DB, error) {
	if database != nil {
		return database, nil
	}

	dbPath := viper.GetString("config_path")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.AutoMigrate(&Device{}, &AESRecordingDevice{}, &PBXConnectionCredentials{}, &AppConfig{}, &UploadRecord{})
	if err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return &DB{
		gormDB: db,
	}, nil
}

// Get all devices that shall be monitored
func (db *DB) GetAllDevices() []Device {
	var devices []Device
	db.gormDB.Find(&devices)
	return devices
}

func (db *DB) GetDeviceById(deviceId int) Device {
	var device Device
	db.gormDB.Where("id = ?", deviceId).First(&device)
	return device
}

// Get all devices that shall be monitored
func (db *DB) GetMonitoredDevices() []Device {
	var devices []Device
	db.gormDB.Where("record_calls = ?", true).Find(&devices)
	return devices
}

// Get all configured AES recording devices
func (db *DB) GetAESRecordingDevices() []AESRecordingDevice {
	var devices []AESRecordingDevice
	db.gormDB.Find(&devices)
	return devices
}

func (db *DB) GetPBXConnectionCredentials() (PBXConnectionCredentials, error) {
	var creds PBXConnectionCredentials
	err := db.gormDB.First(&creds).Error
	return creds, err
}

func (db *DB) GetAppConfig() (AppConfig, error) {
	var config AppConfig
	err := db.gormDB.First(&config).Error
	return config, err
}

func (db *DB) GetRecentUploads() ([]UploadRecord, error) {
	var records []UploadRecord
	err := db.gormDB.Order("created_at desc").Find(&records).Error
	return records, err
}

func (db *DB) Save(value interface{}) (tx *gorm.DB) {
	return db.gormDB.Save(value)
}

func (db *DB) Delete(value interface{}) (tx *gorm.DB) {
	return db.gormDB.Delete(value)
}
