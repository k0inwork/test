package mock

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type GLPIAsset struct {
	ID      string `json:"id"`
	Name    string `json:"name"` // Format: REGION/#seq-POU name
	Long    string `json:"long"`
	Lat     string `json:"lat"`
	Address string `json:"address"`
}

type GLPIProvider struct {
	MockDataPath string
}

func NewGLPIProvider(mockDataPath string) *GLPIProvider {
	return &GLPIProvider{MockDataPath: mockDataPath}
}

func (p *GLPIProvider) GetAssets() []GLPIAsset {
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
				assets := make([]GLPIAsset, 0, len(wrapper.ObjectList))
				for _, obj := range wrapper.ObjectList {
					// We need to re-construct the name format that the sync engine expects:
					// REGION/#seq-POU name
					// Example: MSK/1-ПОУ Rack 1
					fullName := obj.Region + "/" +
						string(rune('0'+obj.SequentialNumber)) + "-" +
						obj.PouType + " " + obj.Name

					assets = append(assets, GLPIAsset{
						ID:      obj.GlpiUUID,
						Name:    fullName,
						Long:    obj.Long,
						Lat:     obj.Lat,
						Address: obj.Address,
					})
				}
				return assets
			}
		}
	}

	// Fallback to static mock data if file loading fails
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
