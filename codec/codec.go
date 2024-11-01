package codec

import "fmt"

var registry map[string]Codec

func init() {
	registry = make(map[string]Codec)

	json := JSONCodec{Indent: "  "}
	Register(".json", &json)

	yaml := YAMLCodec{}
	Register(".yaml", &yaml)
	Register(".yml", &yaml)
}

func Register(ext string, codec Codec) {
	registry[ext] = codec
}

func Get(ext string) (Codec, error) {
	if codec, ok := registry[ext]; ok {
		return codec, nil
	}

	return nil, fmt.Errorf("extension %q has no registered codec", ext)
}

type Codec interface {
	Name() string
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}
