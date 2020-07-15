# GoT

[![GoDoc][godoc-badge]][godoc]

> Pronounced like "goatee".

This package seeks to reduce boilerplate in tests, making it easier to write
more and better tests, particularly in a way that follows best-practices.


## File-driven tests (aka: testdata)

One approach to writing tests, particularly when they are complex to set up, is
to use [file-based test fixtures][dave-cheney-test-fixtures].

Embedding in code is usually a suitable option for light-medium complexity code,
but as things grow more sophisticated, particularly for integration testing and
fuzz testing of non-trivial functions, embedding all that state into code can
become a mess, and (in my experience) less readable the more time has passed.

While opening up files is not hard on it's own, there is usually more to it than
that. You likely need to read the contents, sometimes you decode it as JSON.
Each of these adds more code that distracts from your test. Beyond dealing with
single files, consider reading directories, maybe even recursively. All of that
is just boilerplate, and just serves to distract from the test itself.

This package includes `got.LoadTestData` for loading files on disk into an
annotated struct to eliminate this boilerplate from your own code.

```golang
package mypackage

import (
  "path/filepath"
  "strings"
  "testing"
)

// input.txt
// hello world

// expected.txt
// HELLO WORLD

func TestSomething(t *testing.T) {
  // define the test
  type Test struct {
    Input    string `testdata:"input.txt"`
    Expected string `testdata:"expected.txt"`
  }

  // load the test environment
  var test Test
  got.LoadTestData(t, "testdata", &test)

  // run the code
  actual := strings.ToUpper(test.Input)

  // test the expectations
  if actual != test.Expected {
    t.Fatalf(`expected "%s", got "%s"`, test.Expected, actual)
  }
}
```

Admittedly, this is a contrived example, but it the test is reduced down to
exactly what is being tested, and nothing else.

Beyond this, there is support for reading JSON files and unmarshalling them
automatically and without any additional boilerplate:

```golang
package mypackage

import (
  "path/filepath"
  "reflect"
  "strings"
  "testing"
)

// input.json
// {"a":"hello","b":"world"}

// expected.txt
// {"a":"HELLO","b":"WORLD"}

func TestSomething(t *testing.T) {
  // define the test
  type Test struct {
    Input    map[string]string `testdata:"input.json"`
    Expected map[string]string `testdata:"expected.json"`
  }

  // load the test environment
  var test Test
  got.LoadTestData(t, "testdata", &test)

  // run the code
  actual := make(map[string]string)
  for k, v := range test.Input {
    actual[k] = strings.ToUpper(v)
  }

  // test the expectations
  if !reflect.DeepEqual(actual, test.Expected) {
    t.Fatalf(`expected "%+v", got "%+v"`, test.Expected, actual)
  }
}
```

This library supports decoding `.json` files into structs, maps and other types
via `json.Unmarshal`.

To go a step further, imagine you have a fairly complex output that you want to
test, such as if you're writing some ETL code or operating on binary data.

All of this is fine, but once you have decided your test environment is complex
enough to justify putting the test configuration onto disk, you should probably
be making it easy to write many tests all in the same format, akin to what
[table-driven tests][table-driven-tests] offer for simpler tests.

One approach is to have `testdata/` and have subdirectories for each test, for
example `testdata/some_input_gets_some_output/`. Enter `got.ListSubDirs` which
is just a convenience helper for listing out these types of test cases:

```golang
package mypackage

import (
  "path/filepath"
  "strings"
  "testing"
)

func TestSomething(t *testing.T) {
  // define the test
  type Test struct {
    Input    string `testdata:"input.txt"`
    Expected string `testdata:"expected.txt"`
  }

  // run the exact same test with different inputs (no copy-paste!)
  for _, testName := range got.ListSubDirs(t, "testdata") {
    t.Run(testName, func (t *testing.T) {
      testDir := filepath.Join("testdata", testName)

      // load the test environment
      var test Test
      got.LoadTestData(t, "testdata", &test)

      // run the code
      actual := strings.ToUpper(test.Input)

      // test the expectations
      if actual != test.Expected {
        t.Fatalf(`expected "%s", got "%s"`, test.Expected, actual)
      }
    })
  }
}
```

The next pattern that this library facilitates is [golden files][golden-files],
which are generated when your code is known to be working a particular way, then
saved somewhere that will be read from later when running later tests. An
example of this is HTTP recording, but beyond this can be applied broadly.

Enter `got.SaveTestData`, which is the opposite of `got.LoadTestData` in that it
takes your annotated struct and then saves the data back to disk in the same
format as it would be read. Generally, the only thing you need to treat as
"golden" are the outputs, so we will define 2 structs:

```golang
package mypackage

import (
  "flag"
  "path/filepath"
  "strings"
  "testing"
)

// define a flag to indicate that we should update the golden files
var updateGolden = flag.Bool("update-golden", false, "Update golden test fixtures")

func TestSomething(t *testing.T) {
  // define the test inputs
  type Test struct {
    Input    string `testdata:"input.txt"`
  }

  // define the expectations (separate here so we can save them)
  type Expected struct {
    Output string `testdata:"expected.txt"`
  }

  // run the exact same test with different inputs (no copy-paste!)
  for _, testName := range got.ListSubDirs(t, "testdata") {
    t.Run(testName, func (t *testing.T) {
      testDir := filepath.Join("testdata", testName)

      // load the test environment
      var test Test
      got.LoadTestData(t, "testdata", &test)

      // run the code
      actual := strings.ToUpper(test.Input)

      if *updateGolden {
        // save the outputs back to disk
        // (run again without -update-golden to perform an actual test)
        got.SaveTestData(&Expected{Output: actual})
      } else {
        // test the expectations
        if actual != expected.Output {
          t.Fatalf(`expected "%s", got "%s"`, expected.Output, actual)
        }
      }
    })
  }
}
```

Hopefully this demonstrates a bit of what can be accomplished with file-driven
tests and golden files in particular. GoT is all about getting rid of the
boilerplate that would otherwise obfuscate a complicated test environment. By
doing so, the intention is to make it easier to write more tests, improve test
coverage and overall just make testing easier.

Check out [godoc][godoc] for more information about the API.


[dave-cheney-test-fixtures]: https://dave.cheney.net/2016/05/10/test-fixtures-in-
[golden-files]: https://ieftimov.com/post/testing-in-go-golden-files/
[table-driven-tests]: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
[godoc]: https://godoc.org/github.com/dominicbarnes/got
[godoc-badge]: https://godoc.org/github.com/dominicbarnes/got?status.svg
