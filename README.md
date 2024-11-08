# GoT

[![GoDoc][godoc-badge]][godoc]

> Pronounced like "goatee".

This package is all about making tests easier to write and by improving clarity
through removing boilerplate and code not related to test assertions.

The [Four-Phase Test][four-phase-test] paradigm heavily influences the decisions
made for this library.

## Load: test fixtures as files (aka: testdata)

One approach to writing tests, particularly when they have non-trivial setup, is
to use [file-based test fixtures][dave-cheney-test-fixtures].

Embedding in code is usually a suitable option for light-medium complexity code,
but as things grow more sophisticated, particularly for integration testing and
fuzz testing, embedding all of that state into code gets messy, especially as
time passes.

While opening up files is not difficult on it's own, there can be more to it
(eg: decoding as JSON). Beyond dealing with single files, consider reading
directories (maybe even recursively). Each new line of boilerplate like this
increases the noise-to-signal ratio for the test.

### Working with text (string) and bytes ([]byte)

This package includes `got.Load` for loading files on disk into an annotated
struct to eliminate this boilerplate from your own code.

```golang
package mypackage

import (
  "strings"
  "testing"

  "github.com/dominicbarnes/got/v2"
)

// testdata/input.txt
// hello world

// testdata/expected.txt
// HELLO WORLD

func TestUppercase(t *testing.T) {
  // define test cases
  type Test struct {
    Input    string `testdata:"input.txt"`
    Expected string `testdata:"expected.txt"`
  }

  // load test fixtures
  var test Test
  got.Load(t, "testdata", &test)

  // execute the code under test
  actual := Uppercase(test.Input)

  // perform test assertions
  if actual != test.Expected {
    t.Fatalf(`expected "%s", got "%s"`, test.Expected, actual)
  }
}

// code under test
func Uppercase(input string) string {
  return strings.ToUpper(input)
}
```

While contrived, this demonstates a clear separation between test phases, making
it easier to identify what the test is intending to cover.

Here, simple `string` values are used, but `[]byte` could be used and it would
basically behave as you would expect. (raw file contents, no additional decode)

### Decoding complex types (eg: struct, map, slice)

Taking this to the next logical step, it is also possible for `got.Load` to
unmarshal test fixtures into more sophisticated types (such as a map). The file
extension maps to a codec (eg: JSON, YAML) to perform the decode.

```golang
package mypackage

import (
  "reflect"
  "strings"
  "testing"

  "github.com/dominicbarnes/got/v2"
)

// testdata/input.json
// {
//     "a": "hello",
//     "b": "world"
// }

// testdata/expected.json
// {
//     "a": "HELLO",
//     "b": "WORLD"
// }

func TestUppercaseMap(t *testing.T) {
  // define test cases
  type Test struct {
    Input    map[string]string `testdata:"input.json"`
    Expected map[string]string `testdata:"expected.json"`
  }

  // load test fixtures
  var test Test
  got.LoadTestData(t, "testdata", &test)

  // execute the code under test
  actual := UppercaseMap(test.Input)

  // perform test assertions
  if !reflect.DeepEqual(actual, test.Expected) {
    t.Fatalf(`expected "%+v", got "%+v"`, test.Expected, actual)
  }
}

// code under test
func UppercaseMap(input map[string]string) map[string]string {
  output := make(map[string]string)
  for k, v := range input {
    output[k] = strings.ToUpper(v)
  }
  return output
}
```

Out of the box, this library supports decoding JSON (`.json`) and YAML (`.yml`,
`.yaml`). You can define your own codecs or override the defaults using
`got/codec.Register`.

### Working with dynamic maps of files (explode)

When testing a component that can produce outputs dynamically, or even if just
having a single file for the entire output is undesirable, a `map` type can be
used with the `explode` struct tag option to map to multiple files with a glob.

```golang
package mypackage

import (
  "reflect"
  "strings"
  "testing"

  "github.com/dominicbarnes/got/v2"
)

// testdata/input.json
// {
//   "a": "hello",
//   "b": "world"
// }

// testdata/expected/a.txt
// HELLO

// testdata/expected/a.txt
// WORLD

func TestUppercaseMap(t *testing.T) {
  // define test cases
  type Test struct {
    Input    map[string]string `testdata:"input.json"`
    Expected map[string]string `testdata:"expected/*.txt,explode"`
  }

  // load test fixtures
  var test Test
  got.LoadTestData(t, "testdata", &test)

  // execute the code under test
  actual := UppercaseAsFiles(test.Input)

  // perform test assertions
  if !reflect.DeepEqual(actual, test.Expected) {
    t.Fatalf(`expected "%+v", got "%+v"`, test.Expected, actual)
  }
}

// code under test
func UppercaseAsFiles(input map[string]string) map[string]string {
  output := make(map[string]string)
  for k, v := range input {
    output[fmt.Sprintf("expected/%s.txt", k)] = strings.ToUpper(v)
  }
  return output
}
```

Notice that the `testdata` struct tag uses a glob pattern along with the
`explode` option.

The `Input` map (**not** using `explode`) will look like:

```golang
map[string]string{
  "a": "hello",
  "b": "world",
}
```

The `Expected` map (using `explode`) will look like:

```golang
map[string]string{
  "expected/a.txt": "HELLO",
  "expected/b.txt": "WORLD",
}
```

## Suite: Directory-driven test cases

Consider testing a component with medium-high complexity. Breaking out each case
into manually-defined test functions is workable, but becomes repetitive if the
test setup is always identical.

One approach would be to leverage [table-driven tests][table-driven-tests] to
perform that identical setup within a loop. GoT provides another approach, which
targets a directory and treats each sub-directory there as a separate test case.

```golang
package mypackage

import (
  "strings"
  "testing"

  "github.com/dominicbarnes/got/v2"
)

// testdata/hello-world/input.txt
// hello world

// testdata/hello-world/expected.txt
// HELLO WORLD


// testdata/foo-bar/input.txt
// foo bar

// testdata/foo-bar/expected.txt
// FOO BAR


func TestUppercaseSuite(t *testing.T) {
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

      // execute the code under test
      actual := Uppercase(test.Input)

      // perform test assertions
      if actual != test.Expected {
        t.Fatalf(`expected "%s", got "%s"`, test.Expected, actual)
      }
    },
  }

  // run the test suite: "hello-world" and "foo-bar" each get a sub-test
  suite.Run(t)
}

// code under test
func Uppercase(input string) string {
  return strings.ToUpper(input)
}
```

### Skipping test cases

Sometimes, a test case needs to be disabled temporarily, but deleting it
altogether may not be desirable. To accomplish this, simply rename the directory
to have a ".skip" suffix.

Alternatively, if skipping all but specific tests is desired, add a ".only"
suffix to skip all other test cases.


## Assert: using and updating golden files

In Golang, [golden files][golden-files] are generated when your code is known to
be working as intended, then saved and referenced later to ensure that outputs
do not changed unexpectedly. This is very useful when outputs are difficult to
defined by hand (eg: binary data) or are just large (eg: ETL testing).

`got.Assert` is the companion to `got.Load` in that it takes an annotated struct
but is more focused on writing the data to disk rather than reading it, creating
these "golden files". There are 2 modes of operation here, determined by the
`test.update-golden` flag.

By default, `got.Assert` will compare the input to what already exists on disk,
failing the test if they do not match. When `go test -update-golden` is used,
the input will simply be written to disk, skipping the assertion altogether.


```golang
package mypackage

import (
  "strings"
  "testing"

  "github.com/dominicbarnes/got/v2"
)

// NOTE: no expected.txt files are defined

// testdata/hello-world/input.txt
// hello world

// testdata/foo-bar/input.txt
// foo bar

func TestUppercaseAssert(t *testing.T) {
  // define test inputs
  type Test struct {
    Input string `testdata:"input.txt"`
  }

  // define test expectations
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

      // execute the code under test
      actual := Uppercase(test.Input)

      // perform test assertions
      // 1. tests will fail as expected.txt files are missing (FAIL)
      // 2. add -update-golden and expected.txt files will be written (PASS)
      // 3. tests will pass as long as outputs don't change (PASS)
      got.Assert(&Expected{Output: actual})
    },
  }

  // run the test suite
  suite.Run(t)
}

// code under test
func Uppercase(input string) string {
  return strings.ToUpper(input)
}
```

## RunTestSuite: putting it all together

Using the `RunTestSuite` helper function combines basically every feature above
into an easy-to-grok function call for straightforward test suites.

It uses type parameters (aka: generics) to accept a function with 2 parameters:
`*testing.T` and a test configuration struct (conventionally named `Test`) and
then returning a test assertions struct (conventionally named `Expected`).

The `Test` struct is passed to `Load` automatically and the returned `Expected`
is passed to `Assert` automatically.

```golang
package mypackage

import (
  "strings"
  "testing"

  "github.com/dominicbarnes/got/v2"
)

// testdata/hello-world/input.txt
// hello world

// testdata/hello-world/expected.txt
// HELLO WORLD

// testdata/foo-bar/input.txt
// foo bar

// testdata/foo-bar/expected.txt
// FOO BAR

func TestUppercase(t *testing.T) {
  // define test inputs
  type Test struct {
    Input string `testdata:"input.txt"`
  }

  // define test expectations
  type Expected struct {
    Output string `testdata:"expected.txt"`
  }

  got.RunTestSuite(t, "testdata", func (t *testing.T, tc got.TestCase, test Test) Expected {
    // execute the code under test
    actual := Uppercase(test.Input)

    // return the actual output for assertions 
    return Expected{Output: actual}
  })
}

// code under test
func Uppercase(input string) string {
  return strings.ToUpper(input)
}
```

While contrived, the boilerplate for things like `Load` and `Assert` being
removed really puts the focus on the test itself as much as possible, which is
even more obvious for more sophisticated tests.

Check out [godoc][godoc] for more information about the API.

[dave-cheney-test-fixtures]: https://dave.cheney.net/2016/05/10/test-fixtures-in-
[four-phase-test]: http://xunitpatterns.com/Four%20Phase%20Test.html
[golden-files]: https://ieftimov.com/post/testing-in-go-golden-files/
[table-driven-tests]: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
[godoc]: https://godoc.org/github.com/dominicbarnes/got
[godoc-badge]: https://godoc.org/github.com/dominicbarnes/got?status.svg