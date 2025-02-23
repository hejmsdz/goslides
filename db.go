package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitializeDB(dbPath string, models []interface{}) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("database connection failed")
	}

	db.AutoMigrate(models...)

	return db
}
