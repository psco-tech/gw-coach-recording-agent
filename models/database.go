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

	err = db.AutoMigrate(&Device{})
	if err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return &DB{
		gormDB: db,
	}, nil
}

func (db *DB) GetMonitoredDevices() []Device {
	var devices []Device
	db.gormDB.Where("record_calls = ?", true).Find(&devices)
	return devices
}

func (db *DB) Save(value interface{}) (tx *gorm.DB) {
	return db.gormDB.Save(value)
}
