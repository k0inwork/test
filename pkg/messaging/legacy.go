package messaging

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// MessageFormat defines the wire format for RabbitMQ payloads
type MessageFormat string

const (
	FormatJSON MessageFormat = "json"
	FormatYAML MessageFormat = "yaml"
)

// LegacyEnvelope represents the standard structure expected by legacy Python workers
type LegacyEnvelope struct {
	Command string                 `json:"command" yaml:"command"`
	Args    map[string]interface{} `json:"args" yaml:"args"`
	Meta    map[string]interface{} `json:"meta" yaml:"meta"`
}

// Encode converts a Go struct/map into the specified legacy format
func Encode(v interface{}, format MessageFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.Marshal(v)
	case FormatYAML:
		return yaml.Marshal(v)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Decode parses a legacy format into a Go struct/map
func Decode(data []byte, v interface{}, format MessageFormat) error {
	switch format {
	case FormatJSON:
		return json.Unmarshal(data, v)
	case FormatYAML:
		return yaml.Unmarshal(data, v)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// NewLegacyCommand creates a standard envelope for RabbitMQ commands
func NewLegacyCommand(command string, args map[string]interface{}) *LegacyEnvelope {
	return &LegacyEnvelope{
		Command: command,
		Args:    args,
		Meta:    make(map[string]interface{}),
	}
}
