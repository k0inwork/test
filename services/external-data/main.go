package main

import (
	"log/slog"
	"pum-go/pkg/external"
	"pum-go/pkg/logging"
	"pum-go/services/external-data/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func main() {
	logging.Init("external-data")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "external-data",
		Endpoint:     "http://localhost:8089",
		Capabilities: []string{"external-data", "glpi", "zabbix", "graphql"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
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

	slog.Info("External Data Service starting", "port", 8089)
	if err := r.Run(":8089"); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
