package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
)

const (
	IdentitySvc = "http://localhost:8081"
	ProductSvc  = "http://localhost:8082"
)

func main() {
	logging.Init("frontend")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	// Only base.html exists now
	r.LoadHTMLGlob("services/frontend/templates/base.html")

	r.GET("/", func(c *gin.Context) {
		var users []models.User
		var nodes []models.Product

		respU, err := http.Get(IdentitySvc + "/users")
		if err != nil {
			slog.Error("failed to fetch users", "error", err)
		} else {
			defer respU.Body.Close()
			json.NewDecoder(respU.Body).Decode(&users)
		}

		respN, err := http.Get(ProductSvc + "/nodes")
		if err != nil {
			slog.Error("failed to fetch nodes", "error", err)
		} else {
			defer respN.Body.Close()
			json.NewDecoder(respN.Body).Decode(&nodes)
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"UserCount": len(users),
			"NodeCount": len(nodes),
			"IsIndex":   true,
		})
	})

	r.GET("/nodes", func(c *gin.Context) {
		resp, err := http.Get(ProductSvc + "/nodes")
		var nodes []models.Product
		if err != nil {
			slog.Error("failed to fetch nodes", "error", err)
		} else {
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&nodes)
		}
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Nodes":   nodes,
			"IsNodes": true,
		})
	})

	r.POST("/sync", func(c *gin.Context) {
		slog.Info("sync requested from UI")
		_, err := http.Post(ProductSvc+"/sync", "application/json", nil)
		if err != nil {
			slog.Error("sync request failed", "error", err)
		}
		c.Redirect(http.StatusFound, "/nodes")
	})

	r.GET("/users", func(c *gin.Context) {
		resp, err := http.Get(IdentitySvc + "/users")
		var users []models.User
		if err != nil {
			slog.Error("failed to fetch users", "error", err)
		} else {
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&users)
		}
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Users":   users,
			"IsUsers": true,
		})
	})

	slog.Info("Frontend service starting", "port", 8080)
	r.Run(":8080")
}
