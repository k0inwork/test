package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
)

var (
	GlobalConfig *config.Config
)

type RegisteredService struct {
	Name         string   `json:"name"`
	Endpoint     string   `json:"endpoint"`
	Capabilities []string `json:"capabilities"`
	IsCore       bool     `json:"is_core"`
}

func getServices() ([]RegisteredService, error) {
	resp, err := http.Get(GlobalConfig.Discovery.RegistryURL + "/services")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var services []RegisteredService
	err = json.NewDecoder(resp.Body).Decode(&services)
	return services, err
}

func findServiceByCapability(services []RegisteredService, cap string) string {
	for _, s := range services {
		for _, c := range s.Capabilities {
			if c == cap {
				return s.Endpoint
			}
		}
	}
	return ""
}

func main() {
	logging.Init("frontend")
	cfg, err := config.LoadConfig("system.yaml")
	if err != nil {
		slog.Error("failed to load system.yaml", "error", err)
		panic(err)
	}
	GlobalConfig = cfg
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

	r.LoadHTMLGlob("services/frontend/templates/base.html")

	r.Use(func(c *gin.Context) {
		services, err := getServices()
		if err != nil {
			c.String(http.StatusServiceUnavailable, "Registry Service is Offline")
			c.Abort()
			return
		}

		required := make(map[string]bool)
		for _, name := range GlobalConfig.CoreServices {
			if name == "registry" || name == "frontend" {
				continue
			}
			required[name] = false
		}

		for _, s := range services {
			if _, ok := required[s.Name]; ok {
				required[s.Name] = true
			}
		}

		for name, found := range required {
			if !found {
				c.String(http.StatusServiceUnavailable, fmt.Sprintf("Core Service Offline: %s", name))
				c.Abort()
				return
			}
		}

		c.Set("services", services)
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)

		var users []models.User
		var nodes []models.Product

		idSvc := findServiceByCapability(services, "users")
		prodSvc := findServiceByCapability(services, "nodes")

		if idSvc != "" {
			resp, _ := http.Get(idSvc + "/users")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&users)
			}
		}

		if prodSvc != "" {
			resp, _ := http.Get(prodSvc + "/nodes")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&nodes)
			}
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
			"Services":  services,
		})
	})

	r.GET("/nodes", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		prodSvc := findServiceByCapability(services, "nodes")

		var nodes []models.Product
		if prodSvc != "" {
			resp, _ := http.Get(prodSvc + "/nodes")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&nodes)
			}
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Nodes":    nodes,
			"IsNodes":  true,
			"Services": services,
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
		services := c.MustGet("services").([]RegisteredService)
		prodSvc := findServiceByCapability(services, "sync")
		if prodSvc != "" {
			http.Post(prodSvc+"/sync", "application/json", nil)
		slog.Info("sync requested from UI")
		_, err := http.Post(ProductSvc+"/sync", "application/json", nil)
		if err != nil {
			slog.Error("sync request failed", "error", err)
		}
		c.Redirect(http.StatusFound, "/nodes")
	})

	r.GET("/users", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		idSvc := findServiceByCapability(services, "users")

		var users []models.User
		if idSvc != "" {
			resp, _ := http.Get(idSvc + "/users")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&users)
			}
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Users":    users,
			"IsUsers":  true,
			"Services": services,
		})
	})

	// New: Inventory View
	r.GET("/inventory", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		invSvc := findServiceByCapability(services, "inventory")

		var switches []models.Switch
		if invSvc != "" {
			resp, _ := http.Get(invSvc + "/switches")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&switches)
			}
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Switches":    switches,
			"IsInventory": true,
			"Services":    services,
		})
	})

	// New: Tasks View
	r.GET("/tasks", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		taskSvc := findServiceByCapability(services, "tasks")

		var tasks []models.TaskRecord
		if taskSvc != "" {
			resp, _ := http.Get(taskSvc + "/tasks")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&tasks)
			}
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Tasks":    tasks,
			"IsTasks":  true,
			"Services": services,
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
