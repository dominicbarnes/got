package codec

import (
	"bytes"
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

type YAMLCodec struct {
	Indent int
}

func (c *YAMLCodec) Name() string {
	return "YAML"
}

func (c *YAMLCodec) Marshal(v any) ([]byte, error) {
	if c.Indent > 0 {
		return yamlMarshalIndent(c.Indent, v)
	}

	return yaml.Marshal(v)
}

func (c *YAMLCodec) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

func yamlMarshalIndent(indent int, v any) ([]byte, error) {
	var b bytes.Buffer
	e := yaml.NewEncoder(&b)
	e.SetIndent(indent)
	if err := e.Encode(v); err != nil {
		return nil, fmt.Errorf("yaml encode failed: %w", err)
	}
	return b.Bytes(), nil
}
