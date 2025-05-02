package app

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	analytics "github.com/hejmsdz/api-analytics/analytics/go/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/routers"
)

func newCorsMiddleware(frontendURL string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{frontendURL, "http://localhost:*"}

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
