package got

import "testing"

type tester interface {
	Helper()
	Log(...any)
	Logf(string, ...any)
	Fatal(...any)
	Fatalf(string, ...any)
	Run(string, func(*testing.T)) bool
}
