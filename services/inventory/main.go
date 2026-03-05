package main

import (
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
	if err != nil { panic(err) }
	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{})
}

func main() {
	logging.Init("inventory")
	initDB()
	provider := &external.GraphQLClient{Endpoint: "http://localhost:8089/query"}
	engine := sync.NewSyncEngine(db, provider)
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "inventory",
		Endpoint:     "http://localhost:8083",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "inventory", Endpoints: []string{"/"}},
			{Name: "switches", Endpoints: []string{"/switches"}},
			{Name: "ports", Endpoints: []string{"/ports"}},
			{Name: "sync", Endpoints: []string{"/sync"}},
			{Name: "graphql", Endpoints: []string{"/query"}},
			{Name: "configurable", Endpoints: []string{"/configurable"}},
		},
		IsCore:       false,
		OrderID:      2,
		Menu:         []logging.MenuItem{{Label: "Switches", Path: "/switches"}},
	})

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.GET("/switches", func(c *gin.Context) {
		var switches []models.Switch
		db.Find(&switches)
		c.JSON(200, switches)
	})

	// PDU Mock list (formerly /data/pdu/list)
	r.GET("/pdus", func(c *gin.Context) {
		// Mock response mimicking legacy Django
		c.JSON(200, []map[string]interface{}{
			{"id": "pdu-1", "name": "MSK/1-ПОУ Rack 1", "ip": "10.10.1.5", "status": "Online"},
			{"id": "pdu-2", "name": "SPB/2-ПОУ Rack 2", "ip": "10.20.1.5", "status": "Offline"},
		})
	})

	// IPMI Mock list (formerly /data/ipmi/list)
	r.GET("/ipmi", func(c *gin.Context) {
		// Mock response mimicking legacy Django
		c.JSON(200, []map[string]interface{}{
			{"id": "ipmi-1", "name": "Server-01-IPMI", "ip": "10.10.2.100", "status": "Online"},
		})
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db, Sync: engine}}))
	r.POST("/query", func(c *gin.Context) { srv.ServeHTTP(c.Writer, c.Request) })
	r.GET("/", func(c *gin.Context) { playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request) })

	slog.Info("Inventory starting", "port", 8083)
	r.Run(":8083")
}
