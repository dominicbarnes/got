package codec

import (
	"encoding/json"
	"testing"
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
		testCodec(t, new(JSONCodec), v, json.RawMessage(`{"string":"hello world","integer":42,"boolean":true,"nested":{"string":"foo bar","integer":1234567890}}`))
	})

	t.Run("indent", func(t *testing.T) {
		testCodec(t, &JSONCodec{Indent: "    "}, v, json.RawMessage(`{
    "string": "hello world",
    "integer": 42,
    "boolean": true,
    "nested": {
        "string": "foo bar",
        "integer": 1234567890
    }
}`))
	})
}
