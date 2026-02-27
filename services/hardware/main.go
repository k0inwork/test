package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	logging.Init("hardware")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.POST("/call", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"response": "Hardware call mocked",
		})
	})

	slog.Info("Hardware service starting", "port", 8086)
	r.Run(":8086")
}
