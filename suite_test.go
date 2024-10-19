package got

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunTestSuite(t *testing.T) {
	type Test struct {
		Input string `testdata:"input.txt"`
	}

	type Expected struct {
		Output string `testdata:"expected.txt"`
	}

	RunTestSuite(t, "testdata/suite/assert", func(t *testing.T, tc TestCase, test Test) Expected {
		return Expected{Output: strings.ToUpper(test.Input)}
	})
}

func TestTestSuite(t *testing.T) {
	t.Run("single case", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/single-case",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(&mt, &test)

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

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})

	t.Run("multiple cases", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/multiple-cases",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(&mt, &test)

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

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})

	t.Run("skip", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/skip",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(&mt, &test)

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

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})

	t.Run("only", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir: "testdata/suite/only",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(&mt, &test)

				require.EqualValues(t, "hello world", test.Input)
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name: "test-case-2",
				Only: true,
				Dir:  "testdata/suite/only/test-case-2.only",
			},
		}, cases)

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})

	t.Run("shared dir", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir:       "testdata/suite/shared-dir/cases",
			SharedDir: "testdata/suite/shared-dir/common",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input string `testdata:"input.txt"`
				}

				var test Test
				tc.Load(&mt, &test)

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

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})

	t.Run("shared dir with only", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir:       "testdata/suite/shared-dir-only/cases",
			SharedDir: "testdata/suite/shared-dir-only/common",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input    string `testdata:"input.txt"`
					Expected string `testdata:"expected.txt"`
				}

				var test Test
				tc.Load(&mt, &test)
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name:      "test-case-2",
				Dir:       "testdata/suite/shared-dir-only/cases/test-case-2.only",
				SharedDir: "testdata/suite/shared-dir-only/common/test-case-2",
				Only:      true,
			},
		}, cases)

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})

	t.Run("shared dir with skip", func(t *testing.T) {
		var mt mockT
		var cases []TestCase

		suite := TestSuite{
			Dir:       "testdata/suite/shared-dir-skip/cases",
			SharedDir: "testdata/suite/shared-dir-skip/common",
			TestFunc: func(t *testing.T, tc TestCase) {
				cases = append(cases, tc)

				type Test struct {
					Input    string `testdata:"input.txt"`
					Expected string `testdata:"expected.txt"`
				}

				var test Test
				tc.Load(&mt, &test)
			},
		}

		suite.Run(t)

		require.ElementsMatch(t, []TestCase{
			{
				Name:      "test-case-1",
				Dir:       "testdata/suite/shared-dir-skip/cases/test-case-1",
				SharedDir: "testdata/suite/shared-dir-skip/common/test-case-1",
			},
			{
				Name:      "test-case-3",
				Dir:       "testdata/suite/shared-dir-skip/cases/test-case-3",
				SharedDir: "testdata/suite/shared-dir-skip/common/test-case-3",
			},
		}, cases)

		require.EqualValues(t, mockT{
			helper: true,
		}, mt)
	})
}
