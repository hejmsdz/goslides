package database

import (
	"fmt"

	"github.com/hejmsdz/goslides/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializeDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil || db == nil {
		panic(fmt.Sprintf("database connection failed: %v", err))
	}

	err = models.AutoMigrate(db)
	if err != nil {
		panic(fmt.Sprintf("migrations failed: %v", err))
	}

	return db
}
