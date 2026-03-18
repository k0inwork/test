package external

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// GLPI Models
type GNetworkEquipment struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Model   string `json:"model"`
	IP      string `json:"ip"`
	Status  string `json:"status"`
	Serial  string `json:"serial"`
}

type Gpdu struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Long    string `json:"long"`
	Lat     string `json:"lat"`
	Address string `json:"address"`
	Model   string `json:"model"`
}

type GComputer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Status  string `json:"status"`
}

// Zabbix Models
type ZProblem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Severity int    `json:"severity"`
	Time     string `json:"time"`
}

type ZHost struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	IP       string      `json:"ip"`
	Status   int         `json:"status"`
	Problems []*ZProblem `json:"problems"`
}

// Provider Interface
type Provider interface {
	GetNetworkEquipment(ctx context.Context) ([]*GNetworkEquipment, error)
	GetPDUs(ctx context.Context) ([]*Gpdu, error)
	GetComputers(ctx context.Context) ([]*GComputer, error)
	GetHosts(ctx context.Context) ([]*ZHost, error)
}

// MockProvider implementation
type MockProvider struct {
	MockDataPath string
}

func NewMockProvider(mockDataPath string) *MockProvider {
	return &MockProvider{MockDataPath: mockDataPath}
}

func (p *MockProvider) GetNetworkEquipment(ctx context.Context) ([]*GNetworkEquipment, error) {
	if p.MockDataPath != "" {
		data, err := os.ReadFile(filepath.Join(p.MockDataPath, "switch_list.json"))
		if err == nil {
			var wrapper struct {
				SwitchList []struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					GlpiID      string `json:"glpi_id"`
					LogicalType string `json:"logical_type"`
					Ports       int    `json:"ports"`
					IP          string `json:"ip"`
					Model       string `json:"model"`
					Status      string `json:"status"`
				} `json:"switch_list"`
			}
			if err := json.Unmarshal(data, &wrapper); err == nil {
				res := make([]*GNetworkEquipment, 0, len(wrapper.SwitchList))
				for _, s := range wrapper.SwitchList {
					res = append(res, &GNetworkEquipment{
						ID:     s.ID,
						Name:   s.Name,
						Model:  s.Model,
						IP:     s.IP,
						Status: s.Status,
						Serial: s.GlpiID,
					})
				}
				return res, nil
			}
		}
	}
	return []*GNetworkEquipment{
		{ID: "sw-1", Name: "MSK-SW-01", Model: "Cisco 9300", IP: "10.10.1.1", Status: "Active", Serial: "SN12345"},
		{ID: "sw-2", Name: "SPB-SW-02", Model: "Cisco 9200", IP: "10.20.1.1", Status: "Active", Serial: "SN67890"},
	}, nil
}

func (p *MockProvider) GetPDUs(ctx context.Context) ([]*Gpdu, error) {
	if p.MockDataPath != "" {
		data, err := os.ReadFile(filepath.Join(p.MockDataPath, "products.json"))
		if err == nil {
			var wrapper struct {
				ObjectList []struct {
					ID               int    `json:"id"`
					Name             string `json:"name"`
					Long             string `json:"long"`
					Lat              string `json:"lat"`
					Address          string `json:"address"`
					GlpiUUID         string `json:"glpi_uuid"`
					SequentialNumber int    `json:"sequential_number"`
					PouType          string `json:"pouType"`
					Region           string `json:"region"`
				} `json:"object_list"`
			}
			if err := json.Unmarshal(data, &wrapper); err == nil {
				res := make([]*Gpdu, 0, len(wrapper.ObjectList))
				for _, obj := range wrapper.ObjectList {
					// We need to re-construct the name format that the sync engine expects:
					// REGION/#seq-POU name
					// Example: MSK/1-ПОУ Rack 1
					fullName := obj.Region + "/" +
						string(rune('0'+obj.SequentialNumber)) + "-" +
						obj.PouType + " " + obj.Name

					res = append(res, &Gpdu{
						ID:      obj.GlpiUUID,
						Name:    fullName,
						Long:    obj.Long,
						Lat:     obj.Lat,
						Address: obj.Address,
						Model:   "Generic PDU",
					})
				}
				return res, nil
			}
		}
	}
	return []*Gpdu{
		{ID: "pdu-1", Name: "MSK/1-ПОУ Rack 1", Long: "37.6173", Lat: "55.7558", Address: "Moscow, Red Square", Model: "APC 7921"},
		{ID: "pdu-2", Name: "SPB/2-ПОУ Rack 2", Long: "30.3351", Lat: "59.9343", Address: "St. Petersburg, Palace Square", Model: "APC 7922"},
	}, nil
}

func (p *MockProvider) GetComputers(ctx context.Context) ([]*GComputer, error) {
	if p.MockDataPath != "" {
		data, err := os.ReadFile(filepath.Join(p.MockDataPath, "ipmi_list.json"))
		if err == nil {
			var wrapper struct {
				IpmiList []struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					IP     string `json:"ip"`
					Status string `json:"status"`
				} `json:"ipmi_list"`
			}
			if err := json.Unmarshal(data, &wrapper); err == nil {
				res := make([]*GComputer, 0, len(wrapper.IpmiList))
				for _, c := range wrapper.IpmiList {
					res = append(res, &GComputer{
						ID:     c.ID,
						Name:   c.Name,
						IP:     c.IP,
						Status: c.Status,
					})
				}
				return res, nil
			}
		}
	}
	return []*GComputer{
		{ID: "srv-1", Name: "Server-01", IP: "10.10.1.100", Status: "Online"},
	}, nil
}

func (p *MockProvider) GetHosts(ctx context.Context) ([]*ZHost, error) {
	return []*ZHost{
		{
			ID: "z-1", Name: "MSK-SW-01", IP: "10.10.1.1", Status: 1,
			Problems: []*ZProblem{
				{ID: "e-1", Name: "High CPU usage", Severity: 3, Time: "2024-02-28 10:00:00"},
			},
		},
		{ID: "z-2", Name: "SPB-SW-02", IP: "10.20.1.1", Status: 1, Problems: []*ZProblem{}},
	}, nil
}

// GraphQLClient implementation (to be used by services)
type GraphQLClient struct {
	Endpoint string
}

func (c *GraphQLClient) GetNetworkEquipment(ctx context.Context) ([]*GNetworkEquipment, error) {
	return (&MockProvider{}).GetNetworkEquipment(ctx)
}

func (c *GraphQLClient) GetPDUs(ctx context.Context) ([]*Gpdu, error) {
	return (&MockProvider{}).GetPDUs(ctx)
}

func (c *GraphQLClient) GetComputers(ctx context.Context) ([]*GComputer, error) {
	return (&MockProvider{}).GetComputers(ctx)
}

func (c *GraphQLClient) GetHosts(ctx context.Context) ([]*ZHost, error) {
	return (&MockProvider{}).GetHosts(ctx)
}
