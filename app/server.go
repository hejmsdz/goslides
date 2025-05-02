package app

import (
	"os"

	"github.com/gin-gonic/gin"
	analytics "github.com/hejmsdz/api-analytics/analytics/go/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/routers"
)

func corsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}

func NewApp(container *di.Container) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware)
	r.TrustedPlatform = os.Getenv("TRUSTED_PLATFORM")

	if analyticsKey := os.Getenv("API_ANALYTICS_KEY"); analyticsKey != "" {
		config := analytics.NewConfig()
		config.PrivacyLevel = 1
		r.Use(analytics.AnalyticsWithConfig(analyticsKey, config))
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

	return r
}
