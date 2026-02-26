package main

import (
	"log"
	"net/http"
	"pum-go/pkg/models"
	"pum-go/services/product/graph"
	"pum-go/services/product/sync"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("product.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&models.Product{})
}

func main() {
	initDB()

	r := gin.Default()

	engine := sync.NewSyncEngine(db)

	// REST API
	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(http.StatusOK, products)
	})

	r.POST("/sync", func(c *gin.Context) {
		err := engine.Run()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Sync completed successfully"})
	})

	// GraphQL API
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db, Sync: engine}}))

	r.POST("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	log.Println("Product service starting on :8082")
	r.Run(":8082")
}
