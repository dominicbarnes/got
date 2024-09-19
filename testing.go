package got

import "testing"

//go:generate mockgen -destination=testing_mock.go -package=got . T

type T interface {
	Helper()
	Fatal(...any)
	Fatalf(string, ...any)
	Run(string, func(*testing.T)) bool
}
