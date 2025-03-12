package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializeDB(dsn string, models []interface{}) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("database connection failed")
	}

	db.AutoMigrate(models...)

	return db
}
