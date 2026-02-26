package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/call", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"response": "Hardware call mocked",
		})
	})

	log.Println("Hardware service starting on :8086")
	r.Run(":8086")
}
