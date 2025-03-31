package tests

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/database"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

func SetupTestContainer() *di.Container {
	testDB := database.InitializeDB(os.Getenv("TEST_DATABASE"))
	testContainer := di.NewTestContainer(testDB)

	return testContainer
}

func SetupTestRouter(container *di.Container, registerFuncs ...func(r gin.IRouter, dic *di.Container)) *gin.Engine {
	r := gin.Default()
	for _, register := range registerFuncs {
		register(r, container)
	}
	return r
}

func ClearDatabase(testDB *gorm.DB) {
	unprotectedDB := testDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped()

	for _, model := range models.AllModels {
		unprotectedDB.Delete(model)
	}
}
