package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/routes", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Network routing logic (Mocked)"})
	})

	log.Println("Network service starting on :8084")
	r.Run(":8084")
}
