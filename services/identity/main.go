package main

import (
	"log/slog"
	"net/http"
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
	db.AutoMigrate(&models.User{})
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
		success, role, _ := ldapMock.Authenticate(login.Username, login.Password)
		if success {
			c.JSON(200, gin.H{"username": login.Username, "role": role, "status": "logged_in"})
		} else {
			c.JSON(401, gin.H{"error": "unauth"})
		}
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))
	r.POST("/query", func(c *gin.Context) { srv.ServeHTTP(c.Writer, c.Request) })
	r.GET("/", func(c *gin.Context) { playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request) })

	slog.Info("Identity starting", "port", 8081)
	r.Run(":8081")
}
