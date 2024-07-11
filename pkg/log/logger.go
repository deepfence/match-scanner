package log

var (
	ErrLogger = func(err error) {
		println(err.Error())
	}
	InfoLogger = func(msg string) {
		println(msg)
	}
)
