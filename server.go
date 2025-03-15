package main

import (
	"github.com/gin-gonic/gin"
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

type Server struct {
	container *di.Container
	addr      string
}

func (srv Server) Run() {
	r := gin.Default()
	r.Use(corsMiddleware)

	r.Static("/public", "./public")
	v2 := r.Group("/v2")
	routers.RegisterBootstrapRoutes(v2, srv.container)
	routers.RegisterAuthRoutes(v2, srv.container)
	routers.RegisterSongRoutes(v2, srv.container)
	routers.RegisterDeckRoutes(v2, srv.container)
	routers.RegisterLiturgyRoutes(v2, srv.container)
	routers.RegisterLiveRoutes(v2, srv.container)
	r.Run(srv.addr)
}
