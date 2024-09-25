package codec

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONCodec(t *testing.T) {
	type s struct {
		String  string `json:"string,omitempty"`
		Integer int    `json:"integer,omitempty"`
		Boolean bool   `json:"boolean,omitempty"`
		Nested  *s     `json:"nested,omitempty"`
	}

	v := s{
		String:  "hello world",
		Integer: 42,
		Boolean: true,
		Nested: &s{
			String:  "foo bar",
			Integer: 1234567890,
		},
	}

	t.Run("no indent", func(t *testing.T) {
		raw := `{"string":"hello world","integer":42,"boolean":true,"nested":{"string":"foo bar","integer":1234567890}}`
		testCodec(t, new(JSONCodec), v, json.RawMessage(raw))
	})

	t.Run("indent", func(t *testing.T) {
		raw := `{
    "string": "hello world",
    "integer": 42,
    "boolean": true,
    "nested": {
        "string": "foo bar",
        "integer": 1234567890
    }
}`
		testCodec(t, &JSONCodec{Indent: "    "}, v, json.RawMessage(raw))
	})

	t.Run("max int", func(t *testing.T) {
		c := new(JSONCodec)

		value := map[string]any{"bigint": math.MaxInt64}
		raw := fmt.Sprintf(`{"bigint":%d}`, math.MaxInt64)

		// encode value and ensure we end up with raw
		actual, err := c.Marshal(value)
		require.NoError(t, err)
		require.Equal(t, string(raw), string(actual))

		// decode raw and ensure we end up with a JSON equivalent value
		var decode map[string]any
		require.NoError(t, c.Unmarshal(actual, &decode))
		expected, err := json.Marshal(value)
		require.NoError(t, err)
		actual, err = json.Marshal(decode)
		require.NoError(t, err)
		require.Equal(t, string(expected), string(actual))
	})
}
