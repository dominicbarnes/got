package got

import (
	"fmt"
	"testing"
)

var _ tester = (*mockT)(nil)

type mockT struct {
	helper bool
	failed bool
	logs   []string
}

func (t *mockT) Helper() {
	t.helper = true
}

func (t *mockT) Log(args ...any) {
	t.log(fmt.Sprintln(args...))
}

func (t *mockT) Logf(msg string, args ...any) {
	t.log(fmt.Sprintf(msg, args...))
}

func (t *mockT) Fatal(args ...any) {
	t.Log(args...)
	t.fail()
}

func (t *mockT) Fatalf(msg string, args ...any) {
	t.Logf(msg, args...)
	t.fail()
}

func (t *mockT) Run(name string, fn func(t *testing.T)) bool {
	// TODO
	return true
}

func (t *mockT) log(msg string) {
	t.logs = append(t.logs, msg)
}

func (t *mockT) fail() {
	t.failed = true
}
