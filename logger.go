package application

// Logger is standard logger interface
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}
