package codec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		c, err := Get(".json")
		require.NoError(t, err)
		require.IsType(t, new(JSONCodec), c)
	})

	t.Run("yaml", func(t *testing.T) {
		for _, ext := range []string{".yaml", ".yml"} {
			c, err := Get(ext)
			require.NoError(t, err)
			require.IsType(t, new(YAMLCodec), c)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		c, err := Get(".unknown")
		require.Error(t, err)
		require.Nil(t, c)
	})
}

func testCodec[T any](t *testing.T, c Codec, v1 T, expected []byte) {
	actual, err := c.Marshal(v1)
	require.NoError(t, err)
	require.EqualValues(t, string(expected), string(actual))

	var v2 T
	require.NoError(t, c.Unmarshal(actual, &v2))
	require.EqualValues(t, v1, v2)
}
