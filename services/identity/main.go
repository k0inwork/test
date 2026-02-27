package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/services/identity/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("identity.db"), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		panic(err)
	}

	db.AutoMigrate(&models.User{})

	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		slog.Info("Seeding initial users")
		db.Create(&models.User{Username: "admin", Role: "admin"})
		db.Create(&models.User{Username: "operator", Role: "operator"})
	}
}

func main() {
	logging.Init("identity")
	initDB()

	// Register with Registry
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "identity",
		Endpoint:     "http://localhost:8081",
		Capabilities: []string{"users", "auth"},
		IsCore:       true,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})

	r.POST("/users", func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			slog.Warn("failed to bind user JSON", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&user)
		slog.Info("user created", "username", user.Username)
		c.JSON(http.StatusOK, user)
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))

	r.POST("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	slog.Info("Identity service starting", "port", 8081)
	r.Run(":8081")
}
