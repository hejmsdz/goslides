package tests

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/app"
	"github.com/hejmsdz/goslides/database"
	"github.com/hejmsdz/goslides/di"
	"gorm.io/gorm"
)

type TestEnvironment struct {
	DB *gorm.DB
	T  *testing.T
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
	db := database.InitializeDB(os.Getenv("TEST_DATABASE"))
	ClearDatabase(db)

	return &TestEnvironment{
		DB: db,
		T:  t,
	}
}

type TestCaseEnvironment struct {
	Container *di.Container
	App       *gin.Engine
	DB        *gorm.DB
}

func (e *TestEnvironment) Run(description string, testFunc func(t *testing.T, tce *TestCaseEnvironment)) {
	e.T.Run(description, func(t *testing.T) {
		tx := e.DB.Begin()
		if tx.Error != nil {
			t.Fatalf("Failed to begin transaction: %v", tx.Error)
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r)
			}
		}()

		container := di.NewTestContainer(tx)
		app := app.NewApp(container)
		tce := &TestCaseEnvironment{
			Container: container,
			App:       app,
			DB:        tx,
		}

		testFunc(t, tce)

		if err := tx.Rollback().Error; err != nil {
			t.Errorf("Failed to rollback transaction: %v", err)
		}
	})
}
