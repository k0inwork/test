package main

import (
	"pum-go/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"context"

	"log/slog"
	"pum-go/pkg/external"
	"pum-go/pkg/logging"
	"pum-go/services/external-data/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func main() {
	tp, _ := tracing.InitTracer("external-data")
	defer func() { if err := tp.Shutdown(context.Background()); err != nil { slog.Error("failed to shutdown tracer", "err", err) } }()
	logging.Init("external-data")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "external-data",
		Endpoint:     "http://localhost:8089",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "external-data", Endpoints: []string{"/"}},
			{Name: "glpi", Endpoints: []string{"/glpi"}},
			{Name: "zabbix", Endpoints: []string{"/zabbix"}},
			{Name: "graphql", Endpoints: []string{"/query"}},
		},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(otelgin.Middleware("external-data"))
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	resolver := &graph.Resolver{
		Provider: &external.MockProvider{},
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	r.POST("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	// Unified Problems Endpoint (formerly /products/zabbix_problems/ and /products/nodes-problems/)
	r.GET("/problems", func(c *gin.Context) {
		// Mock response merging problems from Zabbix and internal node status
		c.JSON(200, []map[string]interface{}{
			{
				"id":       "e-1",
				"node":     "MSK-SW-01",
				"severity": 3,
				"time":     "2024-02-28 10:00:00",
				"message":  "High CPU usage",
				"source":   "zabbix",
			},
			{
				"id":       "e-2",
				"node":     "SPB-SW-02",
				"severity": 1,
				"time":     "2024-02-28 10:15:00",
				"message":  "Port 5 link down",
				"source":   "nodes",
			},
		})
	})

	slog.Info("External Data Service starting", "port", 8089)
	if err := r.Run(":8089"); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
