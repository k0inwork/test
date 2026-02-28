package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/external"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"strings"

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
	resp, err := http.Get("http://localhost:8088/services")
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.LoadHTMLGlob("services/frontend/templates/base.html")

	// Middleware to inject services and check core availability
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
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
		})
	})

	r.POST("/sync/nodes", func(c *gin.Context) {
		services, _ := getServices()
		prodSvc := findServiceByCapability(services, "sync")
		if prodSvc != "" {
			http.Post(prodSvc+"/sync", "application/json", nil)
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

	r.GET("/ports", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		invSvc := findServiceByCapability(services, "ports")

		var ports []models.SwitchPort
		if invSvc != "" {
			resp, _ := http.Get(invSvc + "/ports")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&ports)
			}
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Ports":    ports,
			"IsPorts":  true,
			"Services": services,
		})
	})

	r.POST("/sync/inventory", func(c *gin.Context) {
		services, _ := getServices()
		invSvc := findServiceByCapability(services, "sync")
		if invSvc != "" {
			http.Post(invSvc+"/sync", "application/json", nil)
		}
		c.Redirect(http.StatusFound, "/inventory")
	})

	r.GET("/monitoring", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		dataSvc := findServiceByCapability(services, "zabbix")

		var hosts []external.ZHost
		if dataSvc != "" {
			query := `{"query":"{ hosts { id name ip status problems { id name severity time } } }"}`
			resp, err := http.Post(dataSvc+"/query", "application/json", strings.NewReader(query))
			if err == nil {
				defer resp.Body.Close()
				var gqlResp struct {
					Data struct {
						Hosts []external.ZHost `json:"hosts"`
					} `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&gqlResp)
				hosts = gqlResp.Data.Hosts
			}
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Hosts":        hosts,
			"IsMonitoring": true,
			"Services":     services,
		})
	})

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
		})
	})

	slog.Info("Frontend service starting", "port", 8080)
	r.Run(":8080")
}
