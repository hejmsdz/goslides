package main

import (
	"net/http"

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
	Router     *gin.Engine
	HttpServer *http.Server
}

func NewServer(container *di.Container) *Server {
	r := gin.Default()
	r.Use(corsMiddleware)

	r.Static("/public", "./public")
	v2 := r.Group("/v2")
	routers.RegisterBootstrapRoutes(v2, container)
	routers.RegisterAuthRoutes(v2, container)
	routers.RegisterSongRoutes(v2, container)
	routers.RegisterDeckRoutes(v2, container)
	routers.RegisterLiturgyRoutes(v2, container)
	routers.RegisterLiveRoutes(v2, container)

	return &Server{
		Router: r,
	}
}

func (srv *Server) Run(addr string) error {
	srv.HttpServer = &http.Server{
		Addr:    addr,
		Handler: srv.Router.Handler(),
	}
	return srv.HttpServer.ListenAndServe()
}
