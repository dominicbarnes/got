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

	t.Run("string missing file", func(t *testing.T) {
		type TestCase struct {
			Input string `testdata:"does-not-exist"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/text", &actual))
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

	t.Run("bytes missing file", func(t *testing.T) {
		type TestCase struct {
			Input []byte `testdata:"does-not-exist"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/text", &actual))
	})

	t.Run("raw json", func(t *testing.T) {
		type TestCase struct {
			Input json.RawMessage `testdata:"input.json"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/json", &actual))
		expected := TestCase{json.RawMessage("{\n  \"hello\": \"world\"\n}")}
		require.EqualValues(t, expected, actual)
	})

	t.Run("json map", func(t *testing.T) {
		type TestCase struct {
			Input map[string]interface{} `testdata:"input.json"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/json", &actual))
		expected := TestCase{map[string]interface{}{"hello": "world"}}
		require.EqualValues(t, expected, actual)
	})

	t.Run("json map missing file", func(t *testing.T) {
		type TestCase struct {
			Input map[string]interface{} `testdata:"does-not-exist"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/text", &actual))
	})

	t.Run("json interface", func(t *testing.T) {
		type TestCase struct {
			Input interface{} `testdata:"input.json"`
		}
		var actual TestCase
		require.NoError(t, TestData("testdata/json", &actual))
		expected := TestCase{map[string]interface{}{"hello": "world"}}
		require.EqualValues(t, expected, actual)
	})

	t.Run("json interface invalid", func(t *testing.T) {
		type TestCase struct {
			Input interface{} `testdata:"invalid.json"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/json", &actual))
	})

	t.Run("json interface missing file", func(t *testing.T) {
		type TestCase struct {
			Input interface{} `testdata:"does-not-exist"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/json", &actual))
	})

	t.Run("json invalid", func(t *testing.T) {
		type TestCase struct {
			Input map[string]interface{} `testdata:"invalid.json"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/json", &actual))
	})

	t.Run("unsupported slice element type", func(t *testing.T) {
		type TestCase struct {
			Input []bool `testdata:"input.txt"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/text", &actual))
	})

	t.Run("unsupported field type", func(t *testing.T) {
		type TestCase struct {
			Input bool `testdata:"input.txt"`
		}
		var actual TestCase
		require.Error(t, TestData("testdata/text", &actual))
	})

	t.Run("nil", func(t *testing.T) {
		require.NoError(t, TestData("testdata/text", nil))
	})

	t.Run("not pointer", func(t *testing.T) {
		require.Error(t, TestData("testdata/text", struct{}{}))
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
