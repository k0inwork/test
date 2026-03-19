package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/tasklib"
	"pum-go/pkg/tracing"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[*websocket.Conn]string) // conn -> username
	clientsMu sync.Mutex
)

func broadcastMessage(user string, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn, u := range clients {
		if user == "" || user == "all" || u == user {
			err := conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				slog.Error("WebSocket write error", "error", err)
				conn.Close()
				delete(clients, conn)
			}
		}
	}
}

var otelClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

type RegisteredService struct {
	Name         string                           `json:"name"`
	Endpoint     string                           `json:"endpoint"`
	Capabilities []logging.CapabilityRegistration `json:"capabilities"`
	Enabled      bool                             `json:"enabled"`
	IsCore       bool                             `json:"is_core"`
	OrderID      int                              `json:"order_id"`
	Menu         []logging.MenuItem               `json:"menu"`
}

var GlobalConfig *config.Config

func getCommonH(c *gin.Context) gin.H {
	user, _ := c.Get("pum_user")
	role, _ := c.Get("pum_role")
	capsStr, _ := c.Get("pum_caps")
	svcs, _ := c.Get("services")

	var caps []string
	if str, ok := capsStr.(string); ok && str != "" {
		caps = strings.Split(str, ",")
	}
	hasAll := false
	for _, cap := range caps {
		if cap == "*" || cap == "all" {
			hasAll = true
			break
		}
	}

	type NavItem struct {
		Label string
		URL   string
	}
	nav := []NavItem{}

	if svcs != nil {
		for _, s := range svcs.([]RegisteredService) {
			// Check capabilities
			allowed := false

			// Allow if service requires no capabilities
			if len(s.Capabilities) == 0 {
				allowed = true
			} else if hasAll {
				allowed = true
			} else {
				for _, reqCap := range s.Capabilities {
					for _, userCap := range caps {
						if reqCap.Name == userCap && reqCap.Name != "" {
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
		"User":    user,
		"Role":    role,
		"Nav":     nav,
		"IsAdmin": role == "admin",
	}
}

func main() {
	tp, _ := tracing.InitTracer("frontend")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("frontend")

	var err error
	GlobalConfig, err = config.LoadConfig("system.yaml")
	if err != nil {
		slog.Error("failed to load system.yaml", "err", err)
	}

	// Initialize tasklib for recurring tasks
	tasklib.Init("http://localhost:8085")

	r := gin.Default()
	r.Use(otelgin.Middleware("frontend"))
	r.Use(logging.GinMiddleware())
	r.LoadHTMLGlob("services/frontend/templates/*.html")

	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/api/notify" || strings.HasPrefix(c.Request.URL.Path, "/frontend/task/") || c.Request.URL.Path == "/ws" {
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

		req, _ := http.NewRequestWithContext(c.Request.Context(), "GET", "http://localhost:8088/services", nil)
		resp, err := otelClient.Do(req)
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

		// Find identity service endpoint from registry
		identityEndpoint := "http://localhost:8081" // fallback
		reqReg, _ := http.NewRequestWithContext(c.Request.Context(), "GET", "http://localhost:8088/services", nil)
		respReg, err := otelClient.Do(reqReg)
		if err == nil {
			var svcs []RegisteredService
			json.NewDecoder(respReg.Body).Decode(&svcs)
			respReg.Body.Close()
			for _, s := range svcs {
				if s.Name == "identity" {
					identityEndpoint = s.Endpoint
					break
				}
			}
		}

		reqL, _ := http.NewRequestWithContext(c.Request.Context(), "POST", identityEndpoint+"/login", bytes.NewBuffer(data))
		reqL.Header.Set("Content-Type", "application/json")
		resp, err := otelClient.Do(reqL)
		if err != nil || resp.StatusCode != 200 {
			c.HTML(http.StatusUnauthorized, "base.html", gin.H{"IsLogin": true, "Error": "Login failed"})
			return
		}
		var res struct{ Username, Role, Capabilities string }
		json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
		c.SetCookie("pum_user", res.Username, 3600, "/", "", false, true)
		c.SetCookie("pum_role", res.Role, 3600, "/", "", false, true)
		c.SetCookie("pum_caps", res.Capabilities, 3600, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	})

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		user, _ := c.Cookie("pum_user")
		if user == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Error("Failed to upgrade websocket", "error", err)
			return
		}

		clientsMu.Lock()
		clients[conn] = user
		clientsMu.Unlock()

		defer func() {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
		}()

		// Keep connection alive and listen for close
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	// API endpoint for receiving notifications to broadcast
	r.POST("/api/notify", func(c *gin.Context) {
		var payload struct {
			User    string `json:"user"`    // empty or "all" for broadcast
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		broadcastMessage(payload.User, payload.Message)
		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	})

	// Register recurring notification task
	tasklib.RegisterEndpoint(
		"http://localhost:8088", // registry URL
		r,
		"/frontend/task/notify",                      // local webhook path
		"@every 1m",                                  // schedule
		"http://localhost:8080/frontend/task/notify", // target URL reachable by task service
		"system",                                     // username
		"send-timer-notification",                    // operation
		"timer",                                      // object ID
		"Notification",                               // class name
		func(ctx context.Context, payload []byte) error {
			slog.Info("Executing recurring timer notification")

			// Broadcast message to all users
			msg := fmt.Sprintf("System Timer Notification: The time is now %s", time.Now().Format(time.RFC3339))
			broadcastMessage("all", msg)
			return nil
		},
	)

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
		targetUrl := target.Endpoint + path
		if c.Request.URL.RawQuery != "" {
			targetUrl += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), "GET", targetUrl, nil)
		if err != nil {
			c.String(500, "Failed to build request")
			return
		}
		// Forward cookies so backend services know the user
		for _, cookie := range c.Request.Cookies() {
			req.AddCookie(cookie)
		}

		resp, err := otelClient.Do(req)
		if err != nil {
			c.String(502, "Module unreachable")
			return
		}
		defer resp.Body.Close()
		var data interface{}
		json.NewDecoder(resp.Body).Decode(&data)
		h := appendH(getCommonH(c), gin.H{
			"IsModulePage": true,
			"ModuleName":   mod,
			"ModuleData":   data,
		})

		// Check if a specific template exists for this module+path
		importPath := fmt.Sprintf("services/frontend/templates/%s_%s.html", mod, strings.Trim(path, "/"))
		if _, err := os.Stat(importPath); err == nil {
			h["HasCustomTemplate"] = true
			h["CustomTemplateName"] = fmt.Sprintf("content_%s_%s", mod, strings.Trim(path, "/"))
		} else {
			h["HasCustomTemplate"] = false
		}

		c.HTML(200, "base.html", h)
	})

	r.GET("/admin", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.Redirect(http.StatusFound, "/")
			return
		}
		reqA, _ := http.NewRequestWithContext(c.Request.Context(), "GET", "http://localhost:8088/admin/services", nil)
		resp, _ := otelClient.Do(reqA)
		var svcs []RegisteredService
		json.NewDecoder(resp.Body).Decode(&svcs)
		resp.Body.Close()

		// Collect external modules configuration to show in Admin UI
		extModules := make(map[string]map[string]string)
		if GlobalConfig != nil && GlobalConfig.ExternalModules != nil {
			for name, moduleConf := range GlobalConfig.ExternalModules {
				extModules[name] = map[string]string{
					"mode":          moduleConf.Mode,
					"endpoint":      moduleConf.Endpoint,
					"real_endpoint": moduleConf.RealEndpoint,
				}
			}
		}

		c.HTML(200, "base.html", appendH(getCommonH(c), gin.H{
			"IsAdminPage":     true,
			"Modules":         svcs,
			"ExternalModules": extModules,
		}))
	})

	r.POST("/admin/modules/:name/toggle", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.AbortWithStatus(403)
			return
		}
		data, _ := json.Marshal(map[string]bool{"enabled": c.PostForm("enabled") == "true"})
		reqT, _ := http.NewRequestWithContext(c.Request.Context(), "POST", fmt.Sprintf("http://localhost:8088/admin/services/%s/toggle", c.Param("name")), bytes.NewBuffer(data))
		reqT.Header.Set("Content-Type", "application/json")
		otelClient.Do(reqT)
		c.Redirect(http.StatusFound, "/admin")
	})

	r.GET("/admin/users", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.Redirect(http.StatusFound, "/")
			return
		}

		// Resolve identity endpoint from registry first
		identityEndpoint := "http://localhost:8081" // fallback
		reqReg, _ := http.NewRequestWithContext(c.Request.Context(), "GET", "http://localhost:8088/services", nil)
		respReg, err := otelClient.Do(reqReg)
		if err == nil {
			var svcs []RegisteredService
			json.NewDecoder(respReg.Body).Decode(&svcs)
			respReg.Body.Close()
			for _, s := range svcs {
				if s.Name == "identity" {
					identityEndpoint = s.Endpoint
					break
				}
			}
		}

		// Fetch users from identity service
		reqU, _ := http.NewRequestWithContext(c.Request.Context(), "GET", identityEndpoint+"/users", nil)
		for _, cookie := range c.Request.Cookies() {
			reqU.AddCookie(cookie)
		}
		respUsers, err := otelClient.Do(reqU)
		var users interface{}
		if err == nil {
			json.NewDecoder(respUsers.Body).Decode(&users)
			respUsers.Body.Close()
		}

		// Fetch groups from identity service
		reqG, _ := http.NewRequestWithContext(c.Request.Context(), "GET", identityEndpoint+"/groups", nil)
		for _, cookie := range c.Request.Cookies() {
			reqG.AddCookie(cookie)
		}
		respGroups, err := otelClient.Do(reqG)
		var groups interface{}
		if err == nil {
			json.NewDecoder(respGroups.Body).Decode(&groups)
			respGroups.Body.Close()
		}

		// Check if we need to parse groups to an array of objects to render remove buttons correctly.
		var processedUsers []map[string]interface{}
		if users != nil {
			if uArr, ok := users.([]interface{}); ok {
				for _, u := range uArr {
					if uMap, ok := u.(map[string]interface{}); ok {
						if gStr, ok := uMap["groups"].(string); ok && gStr != "" {
							gList := strings.Split(gStr, ", ")
							uMap["groupList"] = gList
						}
						processedUsers = append(processedUsers, uMap)
					}
				}
				users = processedUsers
			}
		}

		c.HTML(200, "base.html", appendH(getCommonH(c), gin.H{
			"IsAdminUsersPage": true,
			"Users":            users,
			"Groups":           groups,
		}))
	})

	r.POST("/admin/users/assign-role", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.AbortWithStatus(403)
			return
		}

		username := c.PostForm("username")
		group := c.PostForm("group")

		if username != "" && group != "" {
			reqReg, _ := http.NewRequestWithContext(c.Request.Context(), "GET", "http://localhost:8088/services", nil)
			respReg, err := otelClient.Do(reqReg)
			if err == nil {
				var svcs []RegisteredService
				json.NewDecoder(respReg.Body).Decode(&svcs)
				respReg.Body.Close()

				var identityEndpoint string
				for _, s := range svcs {
					if s.Name == "identity" {
						identityEndpoint = s.Endpoint
						break
					}
				}
				if identityEndpoint != "" {
					data, _ := json.Marshal(map[string]string{"group": group})
					reqP, _ := http.NewRequestWithContext(c.Request.Context(), "POST", fmt.Sprintf("%s/users/%s/groups", identityEndpoint, url.PathEscape(username)), bytes.NewBuffer(data))
					reqP.Header.Set("Content-Type", "application/json")
					for _, cookie := range c.Request.Cookies() {
						reqP.AddCookie(cookie)
					}
					resp, err := otelClient.Do(reqP)
					if err == nil {
						resp.Body.Close()
					}
				}
			}
		}

		c.Redirect(http.StatusFound, "/admin/users")
	})

	r.POST("/admin/users/remove-role", func(c *gin.Context) {
		role, _ := c.Get("pum_role")
		if role != "admin" {
			c.AbortWithStatus(403)
			return
		}

		username := c.PostForm("username")
		group := c.PostForm("group")

		if username != "" && group != "" {
			reqReg, _ := http.NewRequestWithContext(c.Request.Context(), "GET", "http://localhost:8088/services", nil)
			respReg, err := otelClient.Do(reqReg)
			if err == nil {
				var svcs []RegisteredService
				json.NewDecoder(respReg.Body).Decode(&svcs)
				respReg.Body.Close()

				var identityEndpoint string
				for _, s := range svcs {
					if s.Name == "identity" {
						identityEndpoint = s.Endpoint
						break
					}
				}

				if identityEndpoint != "" {
					reqD, err := http.NewRequestWithContext(c.Request.Context(), "DELETE", fmt.Sprintf("%s/users/%s/groups/%s", identityEndpoint, url.PathEscape(username), url.PathEscape(group)), nil)
					if err == nil {
						for _, cookie := range c.Request.Cookies() {
							reqD.AddCookie(cookie)
						}
						resp, err := otelClient.Do(reqD)
						if err == nil {
							resp.Body.Close()
						}
					}
				}
			}
		}

		c.Redirect(http.StatusFound, "/admin/users")
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
