package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
)

var GlobalConfig *config.Config

type RegisteredService struct {
	Name         string   `json:"name"`
	Endpoint     string   `json:"endpoint"`
	Capabilities []string `json:"capabilities"`
	IsCore       bool     `json:"is_core"`
	Enabled      bool     `json:"enabled"`
}

func hasCap(caps []string, target string) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
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
		if hasCap(s.Capabilities, cap) {
			return s.Endpoint
		}
	}
	return ""
}

func main() {
	logging.Init("frontend")
	cfg, _ := config.LoadConfig("system.yaml")
	GlobalConfig = cfg

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.SetFuncMap(template.FuncMap{
		"HasCap": hasCap,
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
		if c.Request.URL.Path == "/login" {
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
		services, _ := getServices(false)
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
		var res struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		c.SetCookie("pum_user", res.Username, 3600, "/", "", false, true)
		c.SetCookie("pum_role", res.Role, 3600, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("pum_user", "", -1, "/", "", false, true)
		c.SetCookie("pum_role", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		var nodes []models.Product
		prodSvc := findServiceByCapability(services, "nodes")
		if prodSvc != "" {
			resp, _ := http.Get(prodSvc + "/nodes")
			if resp != nil {
				defer resp.Body.Close()
				json.NewDecoder(resp.Body).Decode(&nodes)
			}
		}
		h := getCommonH(c)
		h["NodeCount"] = len(nodes)
		h["IsIndex"] = true
		c.HTML(http.StatusOK, "base.html", h)
	})

	r.GET("/nodes", func(c *gin.Context) {
		services := c.MustGet("services").([]RegisteredService)
		var nodes []models.Product
		prodSvc := findServiceByCapability(services, "nodes")
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
			c.AbortWithStatus(403)
			return
		}
		name := c.Param("name")
		enabled := c.PostForm("enabled") == "true"
		data, _ := json.Marshal(map[string]bool{"enabled": enabled})
		http.Post(fmt.Sprintf("http://localhost:8088/admin/services/%s/toggle", name), "application/json", bytes.NewBuffer(data))
		c.Redirect(http.StatusFound, "/admin")
	})

	slog.Info("Frontend starting", "port", 8080)
	r.Run(":8080")
}
