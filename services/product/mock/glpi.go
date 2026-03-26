// Package mock implements a mock GLPI provider for the product service, loading
// JSON test data to simulate an external asset management system during local dev.
package mock

import ()

type GLPIAsset struct {
	ID      string `json:"id"`
	Name    string `json:"name"` // Format: REGION/#seq-POU name
	Long    string `json:"long"`
	Lat     string `json:"lat"`
	Address string `json:"address"`
}

type GLPIProvider struct{}

func (p *GLPIProvider) GetAssets() []GLPIAsset {
	return []GLPIAsset{
		{
			ID:      "glpi-1",
			Name:    "MSK/1-ПОУ Rack 1",
			Long:    "37.6173",
			Lat:     "55.7558",
			Address: "Moscow, Red Square",
		},
		{
			ID:      "glpi-2",
			Name:    "SPB/2-ПОУ Rack 2",
			Long:    "30.3351",
			Lat:     "59.9343",
			Address: "St. Petersburg, Palace Square",
		},
	}
}
