package got

type logger struct {
	t      tester
	prefix string
}

func (log *logger) Log(msg string, args ...any) {
	log.t.Logf(log.prefix+": "+msg, args...)
}

func (log *logger) WithPrefix(prefix string) *logger {
	return &logger{
		t:      log.t,
		prefix: log.prefix + prefix,
	}
}
