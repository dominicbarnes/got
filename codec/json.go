package codec

import "encoding/json"

type JSONCodec struct {
	Indent string
}

func (c *JSONCodec) Marshal(v any) ([]byte, error) {
	if c.Indent != "" {
		return json.MarshalIndent(v, "", c.Indent)
	} else {
		return json.Marshal(v)
	}
}

func (c *JSONCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
