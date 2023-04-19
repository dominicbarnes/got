package codec

import "fmt"

var registry map[string]Codec

func init() {
	registry = make(map[string]Codec)

	Register(".json", &JSONCodec{
		Indent: "  ",
	})
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
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}
