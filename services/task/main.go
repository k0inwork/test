package main

import (
	"log"
	"net/http"
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
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&models.TaskRecord{})
}

func main() {
	initDB()

	r := gin.Default()

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []models.TaskRecord
		db.Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	log.Println("Task service starting on :8085")
	r.Run(":8085")
}
