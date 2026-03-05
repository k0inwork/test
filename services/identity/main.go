package main

import (
	"log/slog"
	"net/http"
	"strings"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tasklib"
	"pum-go/services/identity/graph"
	"pum-go/services/identity/ldap"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var ldapMock *ldap.MockLDAPProvider

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("identity.db"), &gorm.Config{})
	if err != nil { panic(err) }
	db.AutoMigrate(&models.User{}, &models.Group{}, &models.ActivityLog{})
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 { db.Create(&models.User{Username: "admin", Role: "admin"}) }
}

func main() {
	logging.Init("identity")
	initDB()
	ldapMock = ldap.NewMockLDAPProvider()

	tasklib.Init("http://localhost:8085")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "identity",
		Endpoint:     "http://localhost:8081",
		Capabilities: []string{"users", "auth", "audit"},
		IsCore:       true,
		OrderID:      0,
		Menu:         []logging.MenuItem{
			{Label: "Users", Path: "/users"},
			{Label: "Groups", Path: "/groups"},
			{Label: "Audit Log", Path: "/activitylist"},
		},
	})

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.Preload("Groups").Find(&users)

		// Map response for better table rendering
		var res []map[string]interface{}
		for _, u := range users {
			groupNames := []string{}
			for _, g := range u.Groups {
				groupNames = append(groupNames, g.Name)
			}
			res = append(res, map[string]interface{}{
				"id": u.ID,
				"username": u.Username,
				"role": u.Role,
				"groups": strings.Join(groupNames, ", "),
			})
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/login", func(c *gin.Context) {
		var login struct { Username, Password string }
		if err := c.ShouldBindJSON(&login); err != nil { c.JSON(400, gin.H{"error": "err"}); return }
		success, role, groups, _ := ldapMock.Authenticate(login.Username, login.Password)
		if success {
			// Sync user to DB
			var user models.User
			if err := db.Preload("Groups").Where("username = ?", login.Username).First(&user).Error; err != nil {
				user = models.User{Username: login.Username, Role: role}
				db.Create(&user)
			}

			// Add/Update LDAP groups without clearing manually assigned ones
			for _, g := range groups {
				var group models.Group
				if err := db.Where("name = ?", g.Name).First(&group).Error; err != nil {
					group = models.Group{Name: g.Name, Capabilities: strings.Join(g.Capabilities, ",")}
					db.Create(&group)
				} else {
					group.Capabilities = strings.Join(g.Capabilities, ",")
					db.Save(&group)
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
					db.Model(&user).Association("Groups").Append(&group)
				}
			}

			// Reload user to get all groups (including manually assigned ones)
			db.Preload("Groups").First(&user, user.ID)

			var caps []string
			for _, g := range user.Groups {
				if g.Capabilities != "" {
					caps = append(caps, strings.Split(g.Capabilities, ",")...)
				}
			}

			c.JSON(200, gin.H{
				"username": login.Username,
				"role": role,
				"status": "logged_in",
				"capabilities": strings.Join(caps, ","),
			})
		} else {
			c.JSON(401, gin.H{"error": "unauth"})
		}
	})

	r.GET("/groups", func(c *gin.Context) {
		var groups []models.Group
		db.Preload("Users").Find(&groups)

		// Map response for better table rendering
		var res []map[string]interface{}
		for _, g := range groups {
			userNames := []string{}
			for _, u := range g.Users {
				userNames = append(userNames, u.Username)
			}
			res = append(res, map[string]interface{}{
				"id": g.ID,
				"name": g.Name,
				"capabilities": g.Capabilities,
				"users": strings.Join(userNames, ", "),
			})
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/users/:username/groups", func(c *gin.Context) {
		username := c.Param("username")
		var body struct { GroupName string `json:"group"` }
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		var user models.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var group models.Group
		if err := db.Where("name = ?", body.GroupName).First(&group).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}

		db.Model(&user).Association("Groups").Append(&group)
		c.JSON(http.StatusOK, gin.H{"status": "assigned"})
	})

	// Auditing (formerly /accounts/activitylist/)
	r.GET("/activitylist", func(c *gin.Context) {
		var activities []models.ActivityLog
		db.Order("datetime desc").Limit(100).Find(&activities)

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

	// Register dummy recurring task for integration test
	tasklib.RegisterEndpoint(
		"http://localhost:8088",
		r,
		"/internal/tasks/dummy-identity",
		"@every 10s",
		"http://localhost:8081/internal/tasks/dummy-identity",
		"system",
		"dummy_test_identity",
		"identity-service",
		"IntegrationTest",
		func(payload []byte) error {
			slog.Info("DUMMY_RECURRING_TEST_EXECUTED", "service", "identity")
			return nil
		},
	)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))
	r.POST("/query", func(c *gin.Context) { srv.ServeHTTP(c.Writer, c.Request) })
	r.GET("/", func(c *gin.Context) { playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request) })

	slog.Info("Identity starting", "port", 8081)
	r.Run(":8081")
}
