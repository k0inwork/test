package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var externalModulesURL = "http://localhost:8086" // In a real scenario, discovered via registry or env var

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("network.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.Subnet{}, &models.IPAddress{})
}

// Helper to call external-modules proxy
func callExternalModule(targetIP, command string, param interface{}) error {
	paramJSON, _ := json.Marshal(param)
	reqBody := map[string]interface{}{
		"target_ip": targetIP,
		"command":   command,
		"param":     string(paramJSON),
	}
	reqBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("%s/call", externalModulesURL), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("external module returned status %d", resp.StatusCode)
	}
	return nil
}

type SubnetReq struct {
	ID      string `json:"id"`
	Pools   string `json:"pools"`
	Subnet  string `json:"subnet"`
	Relay   string `json:"relay"`
	Options string `json:"options"`
}

type DHCPReq struct {
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
	Mac      string `json:"mac"`
	SubnetID string `json:"subnetid"`
}

type DNSReq struct {
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}

func setupRouter(dbConn *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(logging.GinMiddleware())
	db = dbConn

	// SUBNETS
	r.GET("/subnets", func(c *gin.Context) {
		var subnets []models.Subnet
		db.Find(&subnets)
		c.JSON(200, subnets)
	})

	r.POST("/subnets", func(c *gin.Context) {
		var req SubnetReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		pools := strings.Split(req.Pools, ",")
		options := strings.Split(req.Options, ",")
		relays := strings.Split(req.Relay, ",")

		poolsList := []map[string]interface{}{}
		for _, p := range pools {
			parts := strings.Split(p, "-")
			if len(parts) == 2 {
				poolsList = append(poolsList, map[string]interface{}{"pool": fmt.Sprintf("%s - %s", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))})
			}
		}

		optionsList := []map[string]interface{}{}
		for _, o := range options {
			parts := strings.Split(o, "=")
			if len(parts) == 2 {
				optionsList = append(optionsList, map[string]interface{}{"name": strings.TrimSpace(parts[0]), "data": strings.TrimSpace(parts[1])})
			}
		}

		relayList := []string{}
		for _, r := range relays {
			if strings.TrimSpace(r) != "" {
				relayList = append(relayList, strings.TrimSpace(r))
			}
		}

		param := map[string]interface{}{
			"subnet": map[string]interface{}{
				"id": req.ID,
				"pools": poolsList,
				"subnet": req.Subnet,
				"relay": map[string]interface{}{"ip-addresses": relayList},
				"option-data": optionsList,
			},
		}

		err := callExternalModule("127.0.0.1", "dhcp subnet add", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.PUT("/subnets", func(c *gin.Context) {
		// Not explicitly in legacy, but good for parity
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.DELETE("/subnets", func(c *gin.Context) {
		var req SubnetReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{"subnetid": req.ID}
		err := callExternalModule("127.0.0.1", "dhcp subnet remove", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})


	// DHCP
	r.GET("/dhcp", func(c *gin.Context) {
		// Mock response simulating RabbitMQ "dhcp host list" RPC result
		c.JSON(200, []map[string]interface{}{
			{"ip": "10.10.1.50", "mac": "00:1A:2B:3C:4D:5E", "hostname": "client-pc-1"},
		})
	})

	r.POST("/dhcp", func(c *gin.Context) {
		var req DHCPReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{
			"host": map[string]interface{}{
				"ip": req.Address,
				"hostname": req.Hostname,
				"mac": req.Mac,
				"subnetid": req.SubnetID,
			},
		}
		err := callExternalModule("127.0.0.1", "dhcp host add", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.PUT("/dhcp", func(c *gin.Context) {
		var req DHCPReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{
			"host": map[string]interface{}{
				"ip": req.Address,
				"hostname": req.Hostname,
				"mac": req.Mac,
				"subnetid": req.SubnetID,
			},
		}
		// First delete existing... handled implicitly by create/overwrite in old backend,
		// but let's just do add since it handles updates in the mock
		err := callExternalModule("127.0.0.1", "dhcp host add", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.DELETE("/dhcp", func(c *gin.Context) {
		var req DHCPReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{
			"host": map[string]interface{}{
				"ip": req.Address,
				"hostname": req.Hostname,
			},
		}
		err := callExternalModule("127.0.0.1", "dhcp host remove", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})


	// DNS
	r.GET("/dns", func(c *gin.Context) {
		// Mock response simulating RabbitMQ "dns host list" RPC result
		c.JSON(200, []map[string]interface{}{
			{"name": "server.local", "ip": "10.10.1.100", "type": "A"},
			{"name": "router.local", "ip": "10.10.1.1", "type": "A"},
		})
	})

	r.POST("/dns", func(c *gin.Context) {
		var req DNSReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{
			"host": map[string]interface{}{
				"ip": req.Address,
				"hostname": req.Hostname,
			},
		}
		err := callExternalModule("127.0.0.1", "dns host add", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.PUT("/dns", func(c *gin.Context) {
		var req DNSReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{
			"host": map[string]interface{}{
				"ip": req.Address,
				"hostname": req.Hostname,
			},
		}
		err := callExternalModule("127.0.0.1", "dns host add", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.DELETE("/dns", func(c *gin.Context) {
		var req DNSReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		param := map[string]interface{}{
			"host": map[string]interface{}{
				"ip": req.Address,
				"hostname": req.Hostname,
			},
		}
		err := callExternalModule("127.0.0.1", "dns host remove", param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func main() {
	logging.Init("network")
	initDB()
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "network",
		Endpoint:     "http://localhost:8084",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "network", Endpoints: []string{"/"}},
			{Name: "ipam", Endpoints: []string{"/ipam"}},
			{Name: "routing", Endpoints: []string{"/routing"}},
		},
		IsCore:       false,
		OrderID:      3,
		Menu:         []logging.MenuItem{{Label: "Subnets", Path: "/subnets"}},
	})

	r := setupRouter(db)

	slog.Info("Network starting", "port", 8084)
	r.Run(":8084")
}
