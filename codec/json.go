package codec

import (
	"bytes"
	"encoding/json"
)

type JSONCodec struct {
	Indent string
}

func (c *JSONCodec) Name() string {
	return "JSON"
}

func (c *JSONCodec) Marshal(v any) ([]byte, error) {
	if c.Indent != "" {
		return json.MarshalIndent(v, "", c.Indent)
	} else {
		return json.Marshal(v)
	}
}

func (c *JSONCodec) Unmarshal(data []byte, v any) error {
	r := bytes.NewBuffer(data)
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}
