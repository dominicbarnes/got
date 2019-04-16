package got

//go:generate mockgen -source testing.go -destination mock_test.go -package got_test TestingT

// TestingT is a wrapper around *testing.T so we can use mocks under test.
type TestingT interface {
	Helper()
	Log(...interface{})
	Logf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}
