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
		Capabilities: []string{"inventory", "switches", "ports", "sync", "graphql", "configurable"},
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

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db, Sync: engine}}))
	r.POST("/query", func(c *gin.Context) { srv.ServeHTTP(c.Writer, c.Request) })
	r.GET("/", func(c *gin.Context) { playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request) })

	slog.Info("Inventory starting", "port", 8083)
	r.Run(":8083")
}
