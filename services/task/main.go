package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("task.db"), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		panic(err)
	}

	db.AutoMigrate(&models.TaskRecord{})
}

func main() {
	logging.Init("task")
	initDB()

	// Register with Registry
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "task",
		Endpoint:     "http://localhost:8085",
		Capabilities: []string{"tasks"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []models.TaskRecord
		db.Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}
