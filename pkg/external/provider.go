package external

import (
	"context"
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
type MockProvider struct{}

func (p *MockProvider) GetNetworkEquipment(ctx context.Context) ([]*GNetworkEquipment, error) {
	return []*GNetworkEquipment{
		{ID: "sw-1", Name: "MSK-SW-01", Model: "Cisco 9300", IP: "10.10.1.1", Status: "Active", Serial: "SN12345"},
		{ID: "sw-2", Name: "SPB-SW-02", Model: "Cisco 9200", IP: "10.20.1.1", Status: "Active", Serial: "SN67890"},
	}, nil
}

func (p *MockProvider) GetPDUs(ctx context.Context) ([]*Gpdu, error) {
	return []*Gpdu{
		{ID: "pdu-1", Name: "MSK/1-ПОУ Rack 1", Long: "37.6173", Lat: "55.7558", Address: "Moscow, Red Square", Model: "APC 7921"},
		{ID: "pdu-2", Name: "SPB/2-ПОУ Rack 2", Long: "30.3351", Lat: "59.9343", Address: "St. Petersburg, Palace Square", Model: "APC 7922"},
	}, nil
}

func (p *MockProvider) GetComputers(ctx context.Context) ([]*GComputer, error) {
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
	// Placeholder for actual GraphQL query logic
	// In a real implementation, this would use a GraphQL client (e.g., shurcooL/graphql)
	// For now, we'll return mock data or call the external-data service
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
