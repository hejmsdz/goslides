package app

import (
	"os"
	"regexp"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/routers"
	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
)

func newCorsMiddleware(frontendURL string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOriginFunc = func(origin string) bool {
		if origin == frontendURL {
			return true
		}

		if strings.HasPrefix(origin, "http://localhost:") {
			return true
		}

		if isInternalIP, _ := regexp.MatchString(`^http://192\.168\.\d{1,3}\.\d{1,3}:`, origin); isInternalIP {
			return true
		}

		return false
	}

	return cors.New(config)
}

func newAnalyticsMiddleware(analyticsKey string) gin.HandlerFunc {
	config := analytics.NewConfig()
	config.PrivacyLevel = 1
	config.GetPath = func(c *gin.Context) string {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/public") && strings.HasSuffix(path, ".pdf") {
			return "/public/_.pdf"
		}

		fullPath := c.FullPath()
		if fullPath != "" {
			return fullPath
		}

		return path
	}

	return analytics.AnalyticsWithConfig(analyticsKey, config)
}

func NewApp(container *di.Container) *gin.Engine {
	r := gin.Default()
	r.Use(newCorsMiddleware(os.Getenv("FRONTEND_URL")))
	r.TrustedPlatform = os.Getenv("TRUSTED_PLATFORM")

	if analyticsKey := os.Getenv("API_ANALYTICS_KEY"); analyticsKey != "" {
		r.Use(newAnalyticsMiddleware(analyticsKey))
	}

	r.Static("/public", "./public")

	r.HEAD("/status", func(c *gin.Context) {
		c.Status(200)
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v2 := r.Group("/v2")
	routers.RegisterBootstrapRoutes(v2, container)
	routers.RegisterAuthRoutes(v2, container)
	routers.RegisterUsersRoutes(v2, container)
	routers.RegisterTeamRoutes(v2, container)
	routers.RegisterSongRoutes(v2, container)
	routers.RegisterDeckRoutes(v2, container)
	routers.RegisterLiturgyRoutes(v2, container)
	routers.RegisterLiveRoutes(v2, container)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "api route not found"})
	})

	return r
}
