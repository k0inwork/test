package main

import (
	"log/slog"
	"net/http"
	"strings"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
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
	db.AutoMigrate(&models.User{}, &models.Group{})
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 { db.Create(&models.User{Username: "admin", Role: "admin"}) }
}

func main() {
	logging.Init("identity")
	initDB()
	ldapMock = ldap.NewMockLDAPProvider()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "identity",
		Endpoint:     "http://localhost:8081",
		Capabilities: []string{"users", "auth"},
		IsCore:       true,
		OrderID:      0,
		Menu:         []logging.MenuItem{{Label: "Users", Path: "/users"}},
	})

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})

	r.POST("/login", func(c *gin.Context) {
		var login struct { Username, Password string }
		if err := c.ShouldBindJSON(&login); err != nil { c.JSON(400, gin.H{"error": "err"}); return }
		success, role, groups, _ := ldapMock.Authenticate(login.Username, login.Password)
		if success {
			// Sync user to DB
			var user models.User
			if err := db.Where("username = ?", login.Username).First(&user).Error; err != nil {
				user = models.User{Username: login.Username, Role: role}
				db.Create(&user)
			}

			// Clear existing groups and re-associate
			db.Model(&user).Association("Groups").Clear()

			var caps []string
			for _, g := range groups {
				var group models.Group
				// Sync group to DB
				if err := db.Where("name = ?", g.Name).First(&group).Error; err != nil {
					group = models.Group{Name: g.Name, Capabilities: strings.Join(g.Capabilities, ",")}
					db.Create(&group)
				} else {
					// Update capabilities if they changed in LDAP
					group.Capabilities = strings.Join(g.Capabilities, ",")
					db.Save(&group)
				}
				db.Model(&user).Association("Groups").Append(&group)
				caps = append(caps, g.Capabilities...)
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
		db.Find(&groups)
		c.JSON(http.StatusOK, groups)
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))
	r.POST("/query", func(c *gin.Context) { srv.ServeHTTP(c.Writer, c.Request) })
	r.GET("/", func(c *gin.Context) { playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request) })

	slog.Info("Identity starting", "port", 8081)
	r.Run(":8081")
}
