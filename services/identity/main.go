package main

import (
	"context"
	"log/slog"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tracing"
	"pum-go/services/identity/ldap"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

var db *gorm.DB
var ldapMock = &ldap.MockLDAPProvider{}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("identity.db"), &gorm.Config{})
	if err == nil {
		db.Use(otelgorm.NewPlugin())
	}
	if err != nil { panic(err) }
	db.AutoMigrate(&models.User{}, &models.Group{}, &models.ActivityLog{})
}

func setupRouter(database *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("identity"))
	r.Use(logging.GinMiddleware())

	r.POST("/login", func(c *gin.Context) {
		var req struct{ Username, Password string }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		ok, role, groups, err := ldapMock.Authenticate(c.Request.Context(), req.Username, req.Password)
		if err != nil || !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		var user models.User
		if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
			user = models.User{Username: req.Username, Role: role}
			database.Create(&user)
		}

		var caps []string
		for _, g := range groups {
			caps = append(caps, g.Capabilities...)
		}

		c.JSON(200, gin.H{
			"username":     req.Username,
			"role":         role,
			"capabilities": strings.Join(caps, ","),
			"status":       "logged_in",
		})
	})

	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		database.Find(&users)
		c.JSON(200, users)
	})

	r.GET("/groups", func(c *gin.Context) {
		var groups []models.Group
		database.Find(&groups)
		c.JSON(200, groups)
	})

	r.POST("/users/:username/groups", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	r.DELETE("/users/:username/groups/:group", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	r.POST("/activitylist", func(c *gin.Context) {
		var log models.ActivityLog
		if err := c.ShouldBindJSON(&log); err == nil {
			database.Create(&log)
		}
		c.JSON(200, gin.H{"status": "success"})
	})

	r.GET("/activitylist", func(c *gin.Context) {
		var logs []models.ActivityLog
		database.Order("datetime desc").Limit(100).Find(&logs)
		c.JSON(200, logs)
	})

	return r
}

func main() {
	tp, _ := tracing.InitTracer("identity")
	defer func() { tp.Shutdown(context.Background()) }()
	logging.Init("identity")
	initDB()

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
	})

	r := setupRouter(db)
	slog.Info("Identity starting", "port", 8081)
	r.Run(":8081")
}
