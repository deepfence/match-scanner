package log

var (
	ErrLogger = func(err error) {
		println(err.Error())
	}

	DebugLogger = func(msg string) {
		println(msg)
	}
)
