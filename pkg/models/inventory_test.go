package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSwitch_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second) // Truncate to second for JSON comparison

	s := Switch{
		ID:          "sw-1",
		Name:        "Core Switch",
		GlpiID:      "123",
		NodeID:      10,
		LogicalType: "cl",
		PortsCount:  48,
		IP:          "192.168.1.1",
		Model:       "Cisco 9300",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var s2 Switch
	err = json.Unmarshal(data, &s2)
	require.NoError(t, err)

	assert.Equal(t, s.ID, s2.ID)
	assert.Equal(t, s.Name, s2.Name)
	assert.Equal(t, s.GlpiID, s2.GlpiID)
	assert.Equal(t, s.NodeID, s2.NodeID)
	assert.Equal(t, s.LogicalType, s2.LogicalType)
	assert.Equal(t, s.PortsCount, s2.PortsCount)
	assert.Equal(t, s.IP, s2.IP)
	assert.Equal(t, s.Model, s2.Model)
	assert.True(t, s.CreatedAt.Equal(s2.CreatedAt))
	assert.True(t, s.UpdatedAt.Equal(s2.UpdatedAt))
}

func TestSwitchPort_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	sp := SwitchPort{
		ID:          "port-1",
		SwitchID:    "sw-1",
		Port:        "Gi1/0/1",
		Description: "Uplink",
		Vlan:        100,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(sp)
	require.NoError(t, err)

	var sp2 SwitchPort
	err = json.Unmarshal(data, &sp2)
	require.NoError(t, err)

	assert.Equal(t, sp.ID, sp2.ID)
	assert.Equal(t, sp.SwitchID, sp2.SwitchID)
	assert.Equal(t, sp.Port, sp2.Port)
	assert.Equal(t, sp.Description, sp2.Description)
	assert.Equal(t, sp.Vlan, sp2.Vlan)
	assert.True(t, sp.CreatedAt.Equal(sp2.CreatedAt))
	assert.True(t, sp.UpdatedAt.Equal(sp2.UpdatedAt))
}

func TestInventory_GORM(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Switch{}, &SwitchPort{})
	require.NoError(t, err)

	s := Switch{
		ID:          "sw-test",
		Name:        "Test Switch",
		GlpiID:      "456",
		NodeID:      20,
		LogicalType: "ac",
		PortsCount:  24,
		IP:          "10.0.0.1",
		Model:       "Aruba 2930F",
	}

	err = db.Create(&s).Error
	require.NoError(t, err)

	var fetchedSwitch Switch
	err = db.First(&fetchedSwitch, "id = ?", "sw-test").Error
	require.NoError(t, err)
	assert.Equal(t, "Test Switch", fetchedSwitch.Name)
	assert.Equal(t, "ac", fetchedSwitch.LogicalType)
	assert.Equal(t, "10.0.0.1", fetchedSwitch.IP)

	sp := SwitchPort{
		ID:          "port-test",
		SwitchID:    "sw-test",
		Port:        "1",
		Description: "Server 1",
		Vlan:        10,
	}

	err = db.Create(&sp).Error
	require.NoError(t, err)

	var fetchedPort SwitchPort
	err = db.First(&fetchedPort, "id = ?", "port-test").Error
	require.NoError(t, err)
	assert.Equal(t, "sw-test", fetchedPort.SwitchID)
	assert.Equal(t, "Server 1", fetchedPort.Description)
}
