package got_test

import (
	"log"
	"strings"
	"testing"

	"github.com/dominicbarnes/got"
)

func ExampleTestData() {
	t := new(testing.T)

	type TestCase struct {
		Input    string `testdata:"input.txt"`
		Expected string `testdata:"expected.txt"`
	}

	var testcase TestCase
	got.LoadTestData(t, "testdata/text", &testcase)

	actual := strings.ToUpper(testcase.Input)
	if actual != testcase.Expected {
		log.Fatalf("expected '%s', but got '%s'", testcase.Expected, actual)
	}
}
