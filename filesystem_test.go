package got_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dominicbarnes/got"
	. "github.com/dominicbarnes/got"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetDirs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockt := NewMockTestingT(ctrl)
	mockt.EXPECT().Helper()

	actual := GetDirs(mockt, "testdata")
	expected := []string{"json", "multiple", "text"}

	require.EqualValues(t, expected, actual)

	t.Run("missing dir", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Fatalf("failed to read testdata dir: %s", "open does-not-exist: no such file or directory")

		GetDirs(mockt, "does-not-exist")
	})
}

func ExampleGetDirs(t *testing.T) {
	for _, testName := range got.GetDirs(t, "testdata") {
		t.Run(testName, func(t *testing.T) {
			testDir := filepath.Join("testdata", testName)

			type TestCase struct {
				Input    string `testdata:"input.txt"`
				Expected string `testdata:"expected.txt"`
			}

			var test TestCase
			LoadTestData(t, testDir, &test)

			actual := strings.ToUpper(test.Input)
			if actual != test.Expected {
				t.Fatalf("actual value '%s' did not match expected value '%s'", actual, test.Expected)
			}
		})
	}
}
