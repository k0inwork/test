package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	envelope := &LegacyEnvelope{
		Command: "test_cmd",
		Args:    map[string]interface{}{"foo": "bar"},
		Meta:    map[string]interface{}{"user": "admin"},
	}

	t.Run("JSON", func(t *testing.T) {
		data, err := Encode(envelope, FormatJSON)
		require.NoError(t, err)
		assert.Contains(t, string(data), "\"command\":\"test_cmd\"")
		assert.Contains(t, string(data), "\"args\":{\"foo\":\"bar\"}")
	})

	t.Run("YAML", func(t *testing.T) {
		data, err := Encode(envelope, FormatYAML)
		require.NoError(t, err)
		assert.Contains(t, string(data), "command: test_cmd")
		assert.Contains(t, string(data), "foo: bar")
	})

	t.Run("Unsupported", func(t *testing.T) {
		_, err := Encode(envelope, MessageFormat("xml"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format: xml")
	})
}

func TestDecode(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		data := []byte("{\"command\":\"test_cmd\",\"args\":{\"foo\":\"bar\"}}")
		var envelope LegacyEnvelope
		err := Decode(data, &envelope, FormatJSON)
		require.NoError(t, err)
		assert.Equal(t, "test_cmd", envelope.Command)
		assert.Equal(t, "bar", envelope.Args["foo"])
	})

	t.Run("YAML", func(t *testing.T) {
		data := []byte("command: test_cmd\nargs:\n  foo: bar")
		var envelope LegacyEnvelope
		err := Decode(data, &envelope, FormatYAML)
		require.NoError(t, err)
		assert.Equal(t, "test_cmd", envelope.Command)
		assert.Equal(t, "bar", envelope.Args["foo"])
	})

	t.Run("Unsupported", func(t *testing.T) {
		var envelope LegacyEnvelope
		err := Decode([]byte("{}"), &envelope, MessageFormat("xml"))
		assert.Error(t, err)
	})

	t.Run("InvalidData", func(t *testing.T) {
		var envelope LegacyEnvelope
		err := Decode([]byte("{invalid"), &envelope, FormatJSON)
		assert.Error(t, err)
	})
}

func TestNewLegacyCommand(t *testing.T) {
	args := map[string]interface{}{"key": "value"}
	cmd := NewLegacyCommand("reboot", args)

	assert.NotNil(t, cmd)
	assert.Equal(t, "reboot", cmd.Command)
	assert.Equal(t, args, cmd.Args)
	assert.NotNil(t, cmd.Meta)
	assert.Len(t, cmd.Meta, 0)
}
