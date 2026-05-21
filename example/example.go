// Package main is a example application for the cors middleware for the gin framework
package main

import (
	"fmt"
	"net/http"
	"time"

	cors "github.com/bytewise43/gin-cors/v2"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.Middleware(&cors.Config{
		AllowedOrigins:   []string{"https://foo.bar"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Length", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	err := r.Run(":8080")
	if err != nil {
		fmt.Printf("Failed to start server due to error: %s", err.Error())
	}
}
