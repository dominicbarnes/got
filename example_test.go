package got_test

import (
	"log"
	"strings"

	"github.com/dominicbarnes/got"
)

func ExampleTestData() {
	type TestCase struct {
		Input    string `testdata:"input.txt"`
		Expected string `testdata:"expected.txt"`
	}

	var testcase TestCase
	if err := got.TestData("testdata/text", &testcase); err != nil {
		log.Fatal(err)
	}

	actual := strings.ToUpper(testcase.Input)
	if actual != testcase.Expected {
		log.Fatalf("expected '%s', but got '%s'", testcase.Expected, actual)
	}
}
