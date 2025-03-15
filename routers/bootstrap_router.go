package routers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
)

func RegisterBootstrapRoutes(r gin.IRouter, dic *di.Container) {
	h := NewBootstrapHandler(dic)

	r.GET("/bootstrap", h.GetBootstrap)
	r.POST("/update_release", h.PostUpdateRelease)
}

type BootstrapHandler struct {
	Bootstrap *dtos.BootstrapResponse
}

func NewBootstrapHandler(dic *di.Container) *BootstrapHandler {
	return &BootstrapHandler{}
}

func (h *BootstrapHandler) UpdateRelease(force bool) {
	if h.Bootstrap != nil && !force {
		return
	}

	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "hejmsdz", "slidesui")

	if err != nil {
		return
	}

	currentVersion, _ := strings.CutPrefix(*release.TagName, "v")
  h.Bootstrap = &dtos.BootstrapResponse{CurrentVersion: currentVersion, AppDownloadURL: *release.HTMLURL}
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
