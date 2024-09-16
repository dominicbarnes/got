package got

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRunTestSuite(t *testing.T) {
	type Test struct {
		Input string `testdata:"input.txt"`
	}

	type Expected struct {
		Output string `testdata:"expected.txt"`
	}

	RunTestSuite(t, "testdata/suite/assert", func(t *testing.T, test Test) Expected {
		t.Helper()
		return Expected{Output: strings.ToUpper(test.Input)}
	})
}

func TestTestSuite(t *testing.T) {
	t.Run("single case", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper().AnyTimes()

		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/single-case",
			TestFunc: func(t *testing.T, tc TestCase) {
				t.Helper()

				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(mockT, &test)

				require.EqualValues(t, "hello world", test.Input)
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name: "test-case-1",
				Dir:  "testdata/suite/single-case/test-case-1",
			},
		}, cases)
	})

	t.Run("multiple cases", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper().AnyTimes()

		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/multiple-cases",
			TestFunc: func(t *testing.T, tc TestCase) {
				t.Helper()

				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(mockT, &test)

				require.EqualValues(t, "hello world", test.Input)
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name: "test-case-1",
				Dir:  "testdata/suite/multiple-cases/test-case-1",
			},
			{
				Name: "test-case-2",
				Dir:  "testdata/suite/multiple-cases/test-case-2",
			},
			{
				Name: "test-case-3",
				Dir:  "testdata/suite/multiple-cases/test-case-3",
			},
		}, cases)
	})

	t.Run("skip", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper().AnyTimes()

		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/skip",
			TestFunc: func(t *testing.T, tc TestCase) {
				t.Helper()

				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(mockT, &test)

				require.EqualValues(t, "hello world", test.Input)
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name: "test-case-1",
				Dir:  "testdata/suite/skip/test-case-1",
			},
			{
				Name: "test-case-3",
				Dir:  "testdata/suite/skip/test-case-3",
			},
		}, cases)
	})

	t.Run("shared dir", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper().AnyTimes()

		var cases []TestCase

		suite := TestSuite{
			Dir:       "testdata/suite/shared-dir/cases",
			SharedDir: "testdata/suite/shared-dir/common",
			TestFunc: func(t *testing.T, tc TestCase) {
				t.Helper()

				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(mockT, &test)

				switch tc.Name {
				case "test-case-1":
					require.EqualValues(t, "override", test.Input)
				case "test-case-2":
					require.EqualValues(t, "hello world", test.Input)
				case "test-case-3":
					require.EqualValues(t, "hello world", test.Input)
				default:
					t.Fatalf("unexpected test case %s", tc.Name)
				}
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name:      "test-case-1",
				Dir:       "testdata/suite/shared-dir/cases/test-case-1",
				SharedDir: "testdata/suite/shared-dir/common/test-case-1",
			},
			{
				Name:      "test-case-2",
				Dir:       "testdata/suite/shared-dir/cases/test-case-2",
				SharedDir: "testdata/suite/shared-dir/common/test-case-2",
			},
			{
				Name:      "test-case-3",
				Dir:       "testdata/suite/shared-dir/cases/test-case-3",
				SharedDir: "testdata/suite/shared-dir/common/test-case-3",
			},
		}, cases)
	})
}
