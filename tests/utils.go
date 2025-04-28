package tests

import (
	"fmt"

	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

func ClearDatabase(testDB *gorm.DB) {
	for _, model := range models.AllModels {
		stmt := &gorm.Statement{DB: testDB}
		stmt.Parse(&model)

		testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", stmt.Schema.Table))
	}
}
