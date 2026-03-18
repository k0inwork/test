package main

import (
	"context"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tracing"
	"pum-go/services/identity/graph"
	"pum-go/services/identity/ldap"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

var db *gorm.DB
var ldapMock *ldap.MockLDAPProvider

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("identity.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{}, &models.Group{}, &models.ActivityLog{})

	// Seed groups
	seedGroups := []models.Group{
		{Name: "tsumadm", Capabilities: "*"},
		{Name: "netadm", Capabilities: "network,ipam,routing"},
		{Name: "devadm", Capabilities: "inventory,switches,ports,products"},
	}
	for _, g := range seedGroups {
		var count int64
		db.Model(&models.Group{}).Where("name = ?", g.Name).Count(&count)
		if count == 0 {
			db.Create(&g)
		}
	}

	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		db.Create(&models.User{Username: "admin", Role: "admin"})
	}
}

func setupRouter(dbConn *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("identity"))
	r.Use(logging.GinMiddleware())

	db = dbConn

	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.WithContext(c.Request.Context()).Preload("Groups").Find(&users)

		// Map response for better table rendering
		var res []map[string]interface{}
		for _, u := range users {
			groupNames := []string{}
			for _, g := range u.Groups {
				groupNames = append(groupNames, g.Name)
			}
			res = append(res, map[string]interface{}{
				"id":       u.ID,
				"username": u.Username,
				"role":     u.Role,
				"groups":   strings.Join(groupNames, ", "),
			})
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/login", func(c *gin.Context) {
		var login struct{ Username, Password string }
		if err := c.ShouldBindJSON(&login); err != nil {
			c.JSON(400, gin.H{"error": "err"})
			return
		}
		success, role, groups, _ := ldapMock.Authenticate(c.Request.Context(), login.Username, login.Password)
		if success {
			// Sync user to DB
			var user models.User
			if err := db.WithContext(c.Request.Context()).Preload("Groups").Where("username = ?", login.Username).First(&user).Error; err != nil {
				user = models.User{Username: login.Username, Role: role}
				db.WithContext(c.Request.Context()).Create(&user)
			}

			// Add/Update LDAP groups without clearing manually assigned ones
			for _, g := range groups {
				var group models.Group
				if err := db.WithContext(c.Request.Context()).Where("name = ?", g.Name).First(&group).Error; err != nil {
					group = models.Group{Name: g.Name, Capabilities: strings.Join(g.Capabilities, ",")}
					db.WithContext(c.Request.Context()).Create(&group)
				} else {
					group.Capabilities = strings.Join(g.Capabilities, ",")
					db.WithContext(c.Request.Context()).Save(&group)
				}

				// Only append if not already present
				hasGroup := false
				for _, ug := range user.Groups {
					if ug.ID == group.ID {
						hasGroup = true
						break
					}
				}
				if !hasGroup {
					db.WithContext(c.Request.Context()).Model(&user).Association("Groups").Append(&group)
				}
			}

			// Reload user to get all groups (including manually assigned ones)
			db.WithContext(c.Request.Context()).Preload("Groups").First(&user, user.ID)

			var caps []string
			for _, g := range user.Groups {
				if g.Capabilities != "" {
					caps = append(caps, strings.Split(g.Capabilities, ",")...)
				}
			}

			c.JSON(200, gin.H{
				"username":     login.Username,
				"role":         role,
				"status":       "logged_in",
				"capabilities": strings.Join(caps, ","),
			})
		} else {
			c.JSON(401, gin.H{"error": "unauth"})
		}
	})

	r.GET("/groups", func(c *gin.Context) {
		var groups []models.Group
		db.WithContext(c.Request.Context()).Preload("Users").Find(&groups)

		// Map response for better table rendering
		var res []map[string]interface{}
		for _, g := range groups {
			userNames := []string{}
			for _, u := range g.Users {
				userNames = append(userNames, u.Username)
			}
			res = append(res, map[string]interface{}{
				"id":           g.ID,
				"name":         g.Name,
				"capabilities": g.Capabilities,
				"users":        strings.Join(userNames, ", "),
			})
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/users/:username/groups", func(c *gin.Context) {
		username := c.Param("username")
		var body struct {
			GroupName string `json:"group"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		var user models.User
		if err := db.WithContext(c.Request.Context()).Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var group models.Group
		if err := db.WithContext(c.Request.Context()).Where("name = ?", body.GroupName).First(&group).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}

		db.WithContext(c.Request.Context()).Model(&user).Association("Groups").Append(&group)
		c.JSON(http.StatusOK, gin.H{"status": "assigned"})
	})

	r.DELETE("/users/:username/groups/:group", func(c *gin.Context) {
		username := c.Param("username")
		groupName := c.Param("group")

		var user models.User
		if err := db.WithContext(c.Request.Context()).Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var group models.Group
		if err := db.WithContext(c.Request.Context()).Where("name = ?", groupName).First(&group).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}

		db.WithContext(c.Request.Context()).Model(&user).Association("Groups").Delete(&group)
		c.JSON(http.StatusOK, gin.H{"status": "removed"})
	})

	// Auditing (formerly /accounts/activitylist/)
	r.POST("/activitylist", func(c *gin.Context) {
		var log models.ActivityLog
		if err := c.ShouldBindJSON(&log); err == nil {
			db.WithContext(c.Request.Context()).Create(&log)
		}
		c.Status(http.StatusOK)
	})

	r.GET("/activitylist", func(c *gin.Context) {
		var activities []models.ActivityLog
		query := db.WithContext(c.Request.Context()).Order("datetime desc").Limit(100)

		if userFilter := c.Query("username"); userFilter != "" {
			query = query.Where("username LIKE ?", "%"+userFilter+"%")
		}
		if methodFilter := c.Query("request_method"); methodFilter != "" {
			query = query.Where("request_method = ?", methodFilter)
		}
		if pathFilter := c.Query("request_url"); pathFilter != "" {
			query = query.Where("request_url LIKE ?", "%"+pathFilter+"%")
		}
		if paramFilter := c.Query("query_params"); paramFilter != "" {
			query = query.Where("query_params LIKE ?", "%"+paramFilter+"%")
		}

		query.Find(&activities)

		// Provide a dummy initial log entry if empty to mimic old behavior
		if len(activities) == 0 {
			dummyLog := models.ActivityLog{
				Username:      "system",
				RequestMethod: "SYSTEM_INIT",
				RequestURL:    "/",
				ResponseCode:  200,
			}
			activities = append(activities, dummyLog)
		}

		c.JSON(http.StatusOK, activities)
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))
	r.POST("/query", func(c *gin.Context) { srv.ServeHTTP(c.Writer, c.Request) })
	r.GET("/", func(c *gin.Context) { playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request) })

	return r
}

func main() {
	tp, _ := tracing.InitTracer("identity")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("identity")
	initDB()
	ldapMock = ldap.NewMockLDAPProvider()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "identity",
		Endpoint: "http://localhost:8081",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "users", Endpoints: []string{"/users"}},
			{Name: "auth", Endpoints: []string{"/login"}},
			{Name: "audit", Endpoints: []string{"/activitylist"}},
		},
		IsCore:  true,
		OrderID: 0,
		Menu: []logging.MenuItem{
			{Label: "Users", Path: "/users"},
			{Label: "Groups", Path: "/groups"},
			{Label: "Audit Log", Path: "/activitylist"},
		},
	})

	r := setupRouter(db)
	slog.Info("Identity starting", "port", 8081)
	r.Run(":8081")
}
