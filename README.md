# GoT

> A collection of go packages with helpers to encourage writing better tests by
> keeping boilerplate to a minimum to make the intent of each test as clear as
> possible.

[![GoDoc][godoc-badge]][godoc]


## testdata

A common pattern when writing tests is to use [file-based test fixtures][dave-cheney-test-fixtures].
This library includes some helper functions for loading files from disk into a
struct to eliminate this boilerplate from your own code.


[dave-cheney-test-fixtures]: https://dave.cheney.net/2016/05/10/test-fixtures-in-go
[godoc]: https://godoc.org/github.com/dominicbarnes/got
[godoc-badge]: https://godoc.org/github.com/dominicbarnes/got?status.svg