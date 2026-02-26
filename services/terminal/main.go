package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/session", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Terminal session mocked"})
	})

	log.Println("Terminal service starting on :8087")
	r.Run(":8087")
}
