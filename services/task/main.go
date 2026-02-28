package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"time"

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

	var count int64
	db.Model(&models.TaskRecord{}).Count(&count)
	if count == 0 {
		for i := 1; i <= 10; i++ {
			status := "SUCCESS"
			if i > 8 {
				status = "PENDING"
			}
			finishedAt := time.Now().Add(-time.Duration(i) * time.Hour)
			var finishedPtr *time.Time
			if status == "SUCCESS" {
				finishedPtr = &finishedAt
			}

			db.Create(&models.TaskRecord{
				Username:      "admin",
				Operation:     fmt.Sprintf("Sync-Job-%02d", i),
				Status:        status,
				StartedAt:     time.Now().Add(-time.Duration(i+1) * time.Hour),
				FinishedAt:    finishedPtr,
				ResultMessage: "Mocked task result",
			})
		}
	}
}

func main() {
	logging.Init("task")
	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "task",
		Endpoint:     "http://localhost:8085",
		Capabilities: []string{"tasks", "async-executor"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []models.TaskRecord
		db.Order("started_at desc").Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	r.POST("/tasks", func(c *gin.Context) {
		var task models.TaskRecord
		if err := c.ShouldBindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		task.Status = "PENDING"
		task.StartedAt = time.Now()
		db.Create(&task)

		go func(t models.TaskRecord) {
			slog.Info("Running async task", "id", t.ID, "op", t.Operation)
			time.Sleep(5 * time.Second)

			finished := time.Now()
			db.Model(&models.TaskRecord{}).Where("id = ?", t.ID).Updates(models.TaskRecord{
				Status:        "SUCCESS",
				FinishedAt:    &finished,
				ResultMessage: "Operation completed successfully",
			})
			slog.Info("Task finished", "id", t.ID)
		}(task)

		c.JSON(http.StatusAccepted, task)
	})

		db.Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}
