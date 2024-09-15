package codec

import (
	"encoding/json"
	"testing"
)

func TestYAMLCodec(t *testing.T) {
	type s struct {
		String  string `yaml:"string,omitempty"`
		Integer int    `yaml:"integer,omitempty"`
		Boolean bool   `yaml:"boolean,omitempty"`
		Nested  *s     `yaml:"nested,omitempty"`
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

	t.Run("indent default", func(t *testing.T) {
		testCodec(t, new(YAMLCodec), v, json.RawMessage(`string: hello world
integer: 42
boolean: true
nested:
    string: foo bar
    integer: 1234567890
`))
	})

	t.Run("indent custom", func(t *testing.T) {
		testCodec(t, &YAMLCodec{Indent: 2}, v, json.RawMessage(`string: hello world
integer: 42
boolean: true
nested:
  string: foo bar
  integer: 1234567890
`))
	})
}
