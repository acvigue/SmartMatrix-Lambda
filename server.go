package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CreateServer() *gin.Engine {
	r := gin.Default()

	corsEnv := os.Getenv("CORS")
	if corsEnv != "" {
		config := cors.DefaultConfig()
		config.AllowOrigins = []string{corsEnv}
		r.Use(cors.New(config))
	}

	r.POST("/schema", PostSchema)
	r.POST("/render", PostRender)
	r.GET("/apps", GetApps)
	r.GET("/apps/:id", GetApp)

	return r
}
