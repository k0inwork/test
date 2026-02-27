package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/services/product/graph"
	"pum-go/services/product/sync"
	"fmt"

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
		slog.Error("failed to connect database", "error", err)
		panic(err)
	}

	db.AutoMigrate(&models.Product{})

	var count int64
	db.Model(&models.Product{}).Count(&count)
	if count == 0 {
		for i := 1; i <= 10; i++ {
			region := "MSK"
			if i > 5 {
				region = "SPB"
			}
			db.Create(&models.Product{
				Name:             fmt.Sprintf("Rack %d", i),
				Region:           region,
				SequentialNumber: i,
				PouType:          "ПОУ",
				Address:          fmt.Sprintf("DataCenter Row %d", i),
			})
		}
	}
}

func main() {
	logging.Init("product")
	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "product",
		Endpoint:     "http://localhost:8082",
		Capabilities: []string{"nodes", "sync"},
		IsCore:       true,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	engine := sync.NewSyncEngine(db)

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(http.StatusOK, products)
	})

	r.POST("/sync", func(c *gin.Context) {
		slog.Info("manual sync triggered")
		err := engine.Run()
		if err != nil {
			slog.Error("sync failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Sync completed successfully"})
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db, Sync: engine}}))

	r.POST("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	slog.Info("Product service starting", "port", 8082)
	r.Run(":8082")
}
