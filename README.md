# GoT

> A collection of go packages with helpers to encourage writing better tests by
> keeping boilerplate to a minimum to make the intent of each test as clear as
> possible.

[![GoDoc][godoc-badge]][godoc]


## testdata

The `testdata` package loads your [test fixtures][dave-cheney-test-fixtures]
into structs to reduce boilerplate in your tests.

```go
type TestCase struct {
  Input    string `testdata:"input.txt"`
  Expected string `testdata:"expected.txt"`
}

// testdata/input.txt
// hello world

// testdata/expected.txt
// HELLO WORLD

func TestStringsToUpper(t *testing.T) {
  var testcase TestCase
  if err := testdata.Load("testdata", &testcase); err != nil {
    t.Fatal(err)
  }
  actual := strings.ToUpper(testcase.Input)
  if actual != testcase.Expected {
    t.Fatalf("actual value '%s' did not match expected value '%s'", actual, expected)
  }
}
```


[dave-cheney-test-fixtures]: https://dave.cheney.net/2016/05/10/test-fixtures-in-go
[godoc]: https://godoc.org/github.com/dominicbarnes/got
[godoc-badge]: https://godoc.org/github.com/dominicbarnes/got?status.svg