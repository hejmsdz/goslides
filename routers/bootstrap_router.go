package routers

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
)

func RegisterBootstrapRoutes(r gin.IRouter, dic *di.Container) {
	h := NewBootstrapHandler(dic)

	r.GET("/bootstrap", h.GetBootstrap)
	r.POST("/update_release", h.PostUpdateRelease)
}

type BootstrapHandler struct {
	Bootstrap dtos.BootstrapResponse
}

func getEnvOrNil(key string) *string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return nil
	}
	return &value
}

func NewBootstrapHandler(dic *di.Container) *BootstrapHandler {
	return &BootstrapHandler{
		Bootstrap: dtos.BootstrapResponse{
			FrontendURL: common.GetFrontendURL(""),
			ContactURL:  getEnvOrNil("CONTACT_URL"),
			SupportURL:  getEnvOrNil("SUPPORT_URL"),
		},
	}
}

func (h *BootstrapHandler) UpdateRelease(force bool) {
	if h.Bootstrap.CurrentVersion != "" && !force {
		return
	}

	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "hejmsdz", "slidesui")

	if err != nil {
		return
	}

	currentVersion, _ := strings.CutPrefix(*release.TagName, "v")
	h.Bootstrap.CurrentVersion = currentVersion

	envUrl := getEnvOrNil("APP_DOWNLOAD_URL")
	if envUrl != nil {
		h.Bootstrap.AppDownloadURL = *envUrl
	} else {
		h.Bootstrap.AppDownloadURL = *release.HTMLURL
	}
}

func (h *BootstrapHandler) GetBootstrap(c *gin.Context) {
	h.UpdateRelease(false)

	c.JSON(http.StatusOK, h.Bootstrap)
}

func (h *BootstrapHandler) PostUpdateRelease(c *gin.Context) {
	go func() {
		time.Sleep(60 * time.Second)
		h.UpdateRelease(true)
	}()

	c.Status(http.StatusNoContent)
}
