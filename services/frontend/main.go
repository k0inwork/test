package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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
	Enabled      bool     `json:"enabled"`
}

func getServices(admin bool) ([]RegisteredService, error) {
	url := "http://localhost:8088/services"
	if admin {
		url = "http://localhost:8088/admin/services"
	}
	resp, err := http.Get(url)
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

	r.SetFuncMap(template.FuncMap{
		"HasCap": func(caps []string, target string) bool {
			for _, c := range caps {
				if c == target {
					return true
				}
			}
			return false
		},
	})

	r.LoadHTMLGlob("services/frontend/templates/base.html")

	getCommonH := func(c *gin.Context) gin.H {
		services, _ := c.Get("services")
		user, _ := c.Get("user")
		role, _ := c.Get("role")
		return gin.H{
			"Services": services,
			"User":     user,
			"Role":     role,
			"IsAdmin":  role == "admin",
		}
	}

	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		user, _ := c.Cookie("pum_user")
		role, _ := c.Cookie("pum_role")
		if user == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Set("role", role)

		services, err := getServices(false)
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

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{"IsLogin": true})
	})

	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		loginData := map[string]string{"username": username, "password": password}
		jsonData, _ := json.Marshal(loginData)
		resp, err := http.Post("http://localhost:8081/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil || resp.StatusCode != http.StatusOK {
			c.HTML(http.StatusUnauthorized, "base.html", gin.H{"IsLogin": true, "Error": "Invalid credentials"})
			return
		}
		defer resp.Body.Close()

		var result struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		c.SetCookie("pum_user", result.Username, 3600, "/", "localhost", false, true)
		c.SetCookie("pum_role", result.Role, 3600, "/", "localhost", false, true)

		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("pum_user", "", -1, "/", "localhost", false, true)
		c.SetCookie("pum_role", "", -1, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/login")
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

		h := getCommonH(c)
		h["UserCount"] = len(users)
		h["NodeCount"] = len(nodes)
		h["IsIndex"] = true
		c.HTML(http.StatusOK, "base.html", h)
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

		h := getCommonH(c)
		h["Nodes"] = nodes
		h["IsNodes"] = true
		c.HTML(http.StatusOK, "base.html", h)
	})

	r.POST("/sync/nodes", func(c *gin.Context) {
		services, _ := getServices(false)
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

		h := getCommonH(c)
		h["Users"] = users
		h["IsUsers"] = true
		c.HTML(http.StatusOK, "base.html", h)
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

		h := getCommonH(c)
		h["Switches"] = switches
		h["IsInventory"] = true
		c.HTML(http.StatusOK, "base.html", h)
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

		h := getCommonH(c)
		h["Ports"] = ports
		h["IsPorts"] = true
		c.HTML(http.StatusOK, "base.html", h)
	})

	r.POST("/sync/inventory", func(c *gin.Context) {
		services, _ := getServices(false)
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

		h := getCommonH(c)
		h["Hosts"] = hosts
		h["IsMonitoring"] = true
		c.HTML(http.StatusOK, "base.html", h)
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

		h := getCommonH(c)
		h["Tasks"] = tasks
		h["IsTasks"] = true
		c.HTML(http.StatusOK, "base.html", h)
	})

	r.GET("/admin", func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.Redirect(http.StatusFound, "/")
			return
		}

		modules, _ := getServices(true)

		h := getCommonH(c)
		h["Modules"] = modules
		h["IsAdminPage"] = true
		c.HTML(http.StatusOK, "base.html", h)
	})

	r.POST("/admin/modules/:name/toggle", func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		name := c.Param("name")
		enabled := c.PostForm("enabled") == "true"

		data, _ := json.Marshal(map[string]bool{"enabled": enabled})
		url := fmt.Sprintf("http://localhost:8088/admin/services/%s/toggle", name)
		http.Post(url, "application/json", bytes.NewBuffer(data))

		c.Redirect(http.StatusFound, "/admin")
	})

	slog.Info("Frontend service starting", "port", 8080)
	r.Run(":8080")
}
