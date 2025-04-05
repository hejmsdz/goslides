package app

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/common"
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

func errorHandlerMiddleware(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	var apiError *common.APIError
	for _, err := range c.Errors {
		if errors.As(err, &apiError) {
			log.Printf("%s: %v", apiError.Message, apiError.InnerError)
			c.AbortWithStatusJSON(apiError.StatusCode, gin.H{
				"error": apiError.Message,
			})
			return
		}
	}
}

func NewApp(container *di.Container) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware)
	r.Use(errorHandlerMiddleware)

	r.Static("/public", "./public")
	v2 := r.Group("/v2")
	routers.RegisterBootstrapRoutes(v2, container)
	routers.RegisterAuthRoutes(v2, container)
	routers.RegisterSongRoutes(v2, container)
	routers.RegisterDeckRoutes(v2, container)
	routers.RegisterLiturgyRoutes(v2, container)
	routers.RegisterLiveRoutes(v2, container)

	return r
}
