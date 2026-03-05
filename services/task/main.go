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
		panic(err)
	}
	db.AutoMigrate(&models.TaskRecord{})
}

func main() {
	logging.Init("task")
	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "task",
		Endpoint:     "http://localhost:8085",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "tasks", Endpoints: []string{"/tasks"}},
			{Name: "async-executor", Endpoints: []string{"/async-executor"}},
		},
		IsCore:       false,
		OrderID:      5,
		Menu: []logging.MenuItem{
			{Label: "Tasks", Path: "/tasks"},
		},
	})

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.GET("/manifest", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"OrderID": 5,
			"Menu": []logging.MenuItem{
				{Label: "Tasks", Path: "/tasks"},
			},
		})
	})

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []models.TaskRecord
		db.Order("started_at desc").Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}
