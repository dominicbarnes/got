package got

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestData(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		type TestCase struct {
			Input string `testdata:"input.txt"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/text", &actual))
		expected := TestCase{"hello world"}
		require.EqualValues(t, expected, actual)
	})

	t.Run("bytes", func(t *testing.T) {
		type TestCase struct {
			Input []byte `testdata:"input.txt"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/text", &actual))
		expected := TestCase{[]byte("hello world")}
		require.EqualValues(t, expected, actual)
	})

	t.Run("json", func(t *testing.T) {
		t.Run("raw", func(t *testing.T) {
			type TestCase struct {
				Input json.RawMessage `testdata:"input.json"`
			}
			var actual TestCase
			require.NoError(t, TestData("testdata/json", &actual))
			expected := TestCase{json.RawMessage("{\n  \"hello\": \"world\"\n}")}
			require.EqualValues(t, expected, actual)
		})

		t.Run("map", func(t *testing.T) {
			type TestCase struct {
				Input map[string]interface{} `testdata:"input.json"`
			}
			var actual TestCase
			require.NoError(t, TestData("testdata/json", &actual))
			expected := TestCase{map[string]interface{}{"hello": "world"}}
			require.EqualValues(t, expected, actual)
		})

		t.Run("interface", func(t *testing.T) {
			type TestCase struct {
				Input interface{} `testdata:"input.json"`
			}
			var actual TestCase
			require.NoError(t, TestData("testdata/json", &actual))
			expected := TestCase{map[string]interface{}{"hello": "world"}}
			require.EqualValues(t, expected, actual)
		})

		t.Run("struct", func(t *testing.T) {
			type TestCase struct {
				Input struct {
					Hello string `json:"hello"`
				} `testdata:"input.json"`
			}
			var actual TestCase
			require.NoError(t, TestData("testdata/json", &actual))
			expected := TestCase{
				Input: struct {
					Hello string `json:"hello"`
				}{
					Hello: "world",
				},
			}
			require.EqualValues(t, expected, actual)
		})

		t.Run("invalid", func(t *testing.T) {
			type TestCase struct {
				Input map[string]interface{} `testdata:"invalid.json"`
			}
			var actual TestCase
			require.Error(t, TestData("testdata/json", &actual))
		})
	})

	t.Run("multiple", func(t *testing.T) {
		type TestCase struct {
			A string `testdata:"a.txt"`
			B []byte `testdata:"b.txt"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/multiple", &actual))
		expected := TestCase{"A", []byte("B")}
		require.EqualValues(t, expected, actual)
	})

	t.Run("nil", func(t *testing.T) {
		require.Error(t, TestData("testdata/text", nil))
	})

	t.Run("non-pointer", func(t *testing.T) {
		require.Error(t, TestData("testdata/text", struct{}{}))
	})

	t.Run("missing file", func(t *testing.T) {
		type TestCase struct {
			Input string `testdata:"does-not-exist"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/text", &actual))
	})

	t.Run("missing struct tag", func(t *testing.T) {
		type TestCase struct {
			Missing string
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/text", &actual))
		require.Empty(t, actual.Missing)
	})

	t.Run("empty struct tag", func(t *testing.T) {
		type TestCase struct {
			Missing string `testdata:""`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/text", &actual))
		require.Empty(t, actual.Missing)
	})

	t.Run("dashed struct tag", func(t *testing.T) {
		type TestCase struct {
			Missing string `testdata:"-"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/text", &actual))
		require.Empty(t, actual.Missing)
	})
}
