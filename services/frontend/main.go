package main

import (
	"encoding/json"
	"log"
	"net/http"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
)

const (
	IdentitySvc = "http://localhost:8081"
	ProductSvc  = "http://localhost:8082"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("services/frontend/templates/*")

	r.GET("/", func(c *gin.Context) {
		// Fetch counts (simplified for demo)
		var users []models.User
		var nodes []models.Product

		respU, _ := http.Get(IdentitySvc + "/users")
		if respU != nil {
			json.NewDecoder(respU.Body).Decode(&users)
		}

		respN, _ := http.Get(ProductSvc + "/nodes")
		if respN != nil {
			json.NewDecoder(respN.Body).Decode(&nodes)
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"UserCount": len(users),
			"NodeCount": len(nodes),
			"IsIndex":   true,
		})
	})

	r.GET("/nodes", func(c *gin.Context) {
		resp, _ := http.Get(ProductSvc + "/nodes")
		var nodes []models.Product
		if resp != nil {
			json.NewDecoder(resp.Body).Decode(&nodes)
		}
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Nodes":   nodes,
			"IsNodes": true,
		})
	})

	r.POST("/sync", func(c *gin.Context) {
		http.Post(ProductSvc+"/sync", "application/json", nil)
		c.Redirect(http.StatusFound, "/nodes")
	})

	r.GET("/users", func(c *gin.Context) {
		resp, _ := http.Get(IdentitySvc + "/users")
		var users []models.User
		if resp != nil {
			json.NewDecoder(resp.Body).Decode(&users)
		}
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Users":   users,
			"IsUsers": true,
		})
	})

	log.Println("Frontend service starting on :8080")
	r.Run(":8080")
}
