package log

var (
	ErrLogger = func(err error) {
		println(err.Error())
	}
)
