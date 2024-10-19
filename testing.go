package got

import "testing"

type tester interface {
	Helper()
	Run(string, func(*testing.T)) bool

	testlogger
}

type testlogger interface {
	Log(...any)
	Logf(string, ...any)
	Fatal(...any)
	Fatalf(string, ...any)
}
