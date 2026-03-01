package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

var GlobalConfig *config.Config

type RegisteredService struct {
	Name         string             `json:"name"`
	Endpoint     string             `json:"endpoint"`
	Capabilities []string           `json:"capabilities"`
	IsCore       bool               `json:"is_core"`
	Enabled      bool               `json:"enabled"`
	OrderID      int                `json:"order_id"`
	Menu         []logging.MenuItem `json:"menu"`
}

func main() {
	logging.Init("frontend")
	cfg, _ := config.LoadConfig("system.yaml")
	GlobalConfig = cfg

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.LoadHTMLGlob("services/frontend/templates/base.html")

	getCommonH := func(c *gin.Context) gin.H {
		services, _ := c.Get("services")
		user, _ := c.Get("pum_user")
		role, _ := c.Get("pum_role")
		capsStr, _ := c.Get("pum_caps")

		var caps []string
		if capsStr != nil {
			caps = strings.Split(capsStr.(string), ",")
		}
		hasAll := false
		for _, cap := range caps {
			if cap == "*" || cap == "all" {
				hasAll = true
				break
			}
		}

		type NavItem struct { Label string; URL string }
		nav := []NavItem{}

		if services != nil {
			for _, s := range services.([]RegisteredService) {
				// Check capabilities
				allowed := false
				if hasAll {
					allowed = true
				} else {
					for _, reqCap := range s.Capabilities {
						for _, userCap := range caps {
							if reqCap == userCap {
								allowed = true
								break
							}
						}
						if allowed {
							break
						}
					}
				}

				if allowed {
					for _, item := range s.Menu {
						nav = append(nav, NavItem{
							Label: item.Label,
							URL:   fmt.Sprintf("/m/%s%s", s.Name, item.Path),
						})
					}
				}
			}
		}

		return gin.H{
			"Nav":     nav,
			"User":    user,
			"Role":    role,
			"IsAdmin": role == "admin",
		}
	}

	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}
		user, _ := c.Cookie("pum_user")
		role, _ := c.Cookie("pum_role")
		caps, _ := c.Cookie("pum_caps")
		if user == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("pum_user", user)
		c.Set("pum_role", role)
		c.Set("pum_caps", caps)

		resp, err := http.Get("http://localhost:8088/services")
		if err == nil {
			var svcs []RegisteredService
			json.NewDecoder(resp.Body).Decode(&svcs)
			resp.Body.Close()
			c.Set("services", svcs)
		}
		c.Next()
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{"IsLogin": true})
	})

	r.POST("/login", func(c *gin.Context) {
		un := c.PostForm("username")
		pw := c.PostForm("password")
		data, _ := json.Marshal(map[string]string{"username": un, "password": pw})
		resp, err := http.Post("http://localhost:8081/login", "application/json", bytes.NewBuffer(data))
		if err != nil || resp.StatusCode != 200 {
			c.HTML(http.StatusUnauthorized, "base.html", gin.H{"IsLogin": true, "Error": "Login failed"})
			return
		}
		var res struct { Username, Role, Capabilities string }
		json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
		c.SetCookie("pum_user", res.Username, 3600, "/", "", false, true)
		c.SetCookie("pum_role", res.Role, 3600, "/", "", false, true)
		c.SetCookie("pum_caps", res.Capabilities, 3600, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("pum_user", "", -1, "/", "", false, true)
		c.SetCookie("pum_role", "", -1, "/", "", false, true)
		c.SetCookie("pum_caps", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "base.html", appendH(getCommonH(c), gin.H{"IsIndex": true}))
	})

	r.GET("/m/:module/*path", func(c *gin.Context) {
		mod := c.Param("module")
		path := c.Param("path")
		svcs := c.MustGet("services").([]RegisteredService)
		var target *RegisteredService
		for _, s := range svcs {
			if s.Name == mod {
				target = &s
				break
			}
		}
		if target == nil {
			c.String(404, "Module not found")
			return
		}
		resp, err := http.Get(target.Endpoint + path)
		if err != nil {
			c.String(502, "Module unreachable")
			return
		}
		defer resp.Body.Close()
		var data interface{}
		json.NewDecoder(resp.Body).Decode(&data)
		c.HTML(200, "base.html", appendH(getCommonH(c), gin.H{
			"IsModulePage": true,
			"ModuleName":   mod,
			"ModuleData":   data,
		}))
	})

	r.GET("/admin", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.Redirect(http.StatusFound, "/")
			return
		}
		resp, _ := http.Get("http://localhost:8088/admin/services")
		var svcs []RegisteredService
		json.NewDecoder(resp.Body).Decode(&svcs)
		resp.Body.Close()
		c.HTML(200, "base.html", appendH(getCommonH(c), gin.H{"IsAdminPage": true, "Modules": svcs}))
	})

	r.POST("/admin/modules/:name/toggle", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.AbortWithStatus(403)
			return
		}
		data, _ := json.Marshal(map[string]bool{"enabled": c.PostForm("enabled") == "true"})
		http.Post(fmt.Sprintf("http://localhost:8088/admin/services/%s/toggle", c.Param("name")), "application/json", bytes.NewBuffer(data))
		c.Redirect(http.StatusFound, "/admin")
	})

	slog.Info("Frontend starting", "port", 8080)
	r.Run(":8080")
}

func appendH(h1, h2 gin.H) gin.H {
	for k, v := range h2 {
		h1[k] = v
	}
	return h1
}
