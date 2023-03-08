package application

// Logger is standard logger interface
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type emptyLogger struct{}

func (e *emptyLogger) Fatal(v ...interface{})                 {}
func (e *emptyLogger) Fatalf(format string, v ...interface{}) {}
func (e *emptyLogger) Print(v ...interface{})                 {}
func (e *emptyLogger) Println(v ...interface{})               {}
func (e *emptyLogger) Printf(format string, v ...interface{}) {}
