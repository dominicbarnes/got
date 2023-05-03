# GoT

[![GoDoc][godoc-badge]][godoc]

> Pronounced like "goatee".

This package seeks to reduce boilerplate in tests, making it easier to write more
and better tests, particularly in a way that follows best-practices.

## File-driven tests (aka: testdata)

One approach to writing tests, particularly when they are complex to set up, is
to use [file-based test fixtures][dave-cheney-test-fixtures].

Embedding in code is usually a suitable option for light-medium complexity code,
but as things grow more sophisticated, particularly for integration testing and
fuzz testing of non-trivial functions, embedding all that state into code can
become a mess, and (in my experience) less readable the more time has passed.

While opening up files is not hard on it's own, there is usually more to it than
that. You likely need to read the contents, sometimes you decode it as JSON. Each
of these adds more code that distracts from your test. Beyond dealing with single
files, consider reading directories, maybe even recursively. All of that is just
boilerplate, and just serves to distract from the test itself.

### Load fixtures for a single test

This package includes `got.Load` for loading files on disk into an annotated
struct to eliminate this boilerplate from your own code.

```golang
package mypackage

import (
  "path/filepath"
  "strings"
  "testing"
)

// testdata/input.txt
// hello world

// testdata/expected.txt
// HELLO WORLD

func TestSomething(t *testing.T) {
  // define test case
  type Test struct {
    Input    string `testdata:"input.txt"`
    Expected string `testdata:"expected.txt"`
  }

  // load test fixtures
  var test Test
  got.Load(t, "testdata", &test)

  // run the code
  actual := strings.ToUpper(test.Input)

  // run test assertions
  if actual != test.Expected {
    t.Fatalf(`expected "%s", got "%s"`, test.Expected, actual)
  }
}
```

This is a contrived example, but the test code itself is pretty clear, without
much distraction.

### Load test fixtures into a complex type (eg: map, struct, slice)

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
  // define test cases
  type Test struct {
    Input    map[string]string `testdata:"input.json"`
    Expected map[string]string `testdata:"expected.json"`
  }

  // load test fixtures
  var test Test
  got.LoadTestData(t, "testdata", &test)

  // run the code
  actual := make(map[string]string)
  for k, v := range test.Input {
    actual[k] = strings.ToUpper(v)
  }

  // run test assertions
  if !reflect.DeepEqual(actual, test.Expected) {
    t.Fatalf(`expected "%+v", got "%+v"`, test.Expected, actual)
  }
}
```

Out of the box, this library supports decoding `.json`, `.yml` and `.yaml` files
into structs, maps and other types automatically. You can define your own codecs
using `codec.Register`.

### Running a test for each directory (aka: suite)

To go a step further, imagine you have a fairly complex output that you want to
test, such as if you're writing some ETL code or operating on binary data.

All of this is fine, but once you have decided your test environment is complex
enough to justify putting the test configuration onto disk, you should probably
be making it easy to write many tests all in the same format, akin to what
[table-driven tests][table-driven-tests] offer for simpler tests.

One approach is to have `testdata/` and have subdirectories for each test, for
example `testdata/some_input_gets_some_output/`. Enter `got.TestSuite` which is
just a helper for executing a series of tests using sub-directories.

```golang
package mypackage

import (
  "path/filepath"
  "strings"
  "testing"
)

func TestSomething(t *testing.T) {
  // define test cases
  type Test struct {
    Input    string `testdata:"input.txt"`
    Expected string `testdata:"expected.txt"`
  }

  // define test suite
  suite := got.TestSuite{
    Dir: "testdata",
    TestFunc: func (t *testing.T, c got.TestCase) {
      // load test fixtures
      var test Test
      c.Load(t, &test)

      // run the code
      actual := strings.ToUpper(test.Input)

      // run test assertions
      if actual != test.Expected {
        t.Fatalf(`expected "%s", got "%s"`, test.Expected, actual)
      }
    },
  }

  // run the test suite
  suite.Run(t)
}
```

### Using golden files

The next pattern that this library facilitates is [golden files][golden-files],
which are generated when your code is known to be working a particular way, then
saved somewhere that will be read from later when running later tests. An
example of this is HTTP recording, but the possibilities are quite broad.

Enter `got.Assert`, which is the companion to `got.Load` in that it takes your
annotated struct and then saves the data back to disk in the same format as it
would be read. Generally, the only thing you need to treat as "golden" are the
outputs, so we will define 2 structs:

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
    Input string `testdata:"input.txt"`
  }

  // define the expectations
  type Expected struct {
    Output string `testdata:"expected.txt"`
  }

  // define test suite
  suite := got.TestSuite{
    Dir: "testdata",
    TestFunc: func (t *testing.T, c got.TestCase) {
      // load test fixtures
      var test Test
      c.Load(t, &test)

      // run the code
      actual := strings.ToUpper(test.Input)

      // by default, run test assertions
      // when -update-golden is used, save the golden outputs to disk
      got.Assert(&Expected{Output: actual})
    },
  }

  // run the test suite
  suite.Run(t)
}
```

### Skipping test cases

Sometimes, a test case needs to be disabled temporarily, but deleting it
altogether may not be desirable. To accomplish this, simply rename the directory
to have a ".skip" suffix.

---

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
