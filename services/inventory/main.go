package main

import (
	"fmt"
	"log/slog"
	"pum-go/pkg/external"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/services/inventory/graph"
	"pum-go/services/inventory/sync"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("inventory.db"), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		panic(err)
	}

	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{})

	var count int64
	db.Model(&models.Switch{}).Count(&count)
	if count == 0 {
		for i := 1; i <= 5; i++ {
			swID := fmt.Sprintf("sw-%d", i)
			sw := models.Switch{
				ID: swID, Name: fmt.Sprintf("Switch-%02d", i),
				IP: fmt.Sprintf("10.10.0.%d", i),
				Model: "Cisco Catalyst 9300",
				LogicalType: "cl", PortsCount: 48,
			}
			db.Create(&sw)

			// Create 3 ports for each switch
			for p := 1; p <= 3; p++ {
				db.Create(&models.SwitchPort{
					ID: fmt.Sprintf("p-%d-%d", i, p),
					SwitchID: swID,
					Port: fmt.Sprintf("%s:Port %d", sw.Name, p),
					Vlan: 10 * p,
				})
			}
		}
	}
}

func main() {
	logging.Init("inventory")
	initDB()

	provider := &external.GraphQLClient{Endpoint: "http://localhost:8089/query"}
	engine := sync.NewSyncEngine(db, provider)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "inventory",
		Endpoint:     "http://localhost:8083",
		Capabilities: []string{"inventory", "switches", "ports", "sync", "graphql"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	// REST API
	r.GET("/switches", func(c *gin.Context) {
		var switches []models.Switch
		db.Find(&switches)
		c.JSON(200, switches)
	})

	r.GET("/ports", func(c *gin.Context) {
		var ports []models.SwitchPort
		db.Find(&ports)
		c.JSON(200, ports)
	})

	r.POST("/sync", func(c *gin.Context) {
		slog.Info("manual inventory sync triggered")
		err := engine.Run()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "Inventory sync completed"})
	})

	// GraphQL API
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db, Sync: engine}}))

	r.POST("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	slog.Info("Inventory service starting", "port", 8083)
	r.Run(":8083")
}
