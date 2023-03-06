package application

import "errors"

var (
	ErrWrongState = errors.New("the method called does not match the state")

	ErrInitFailure          = errors.New("initialization error")
	ErrInitTimeout          = errors.New("initialization timeout")
	ErrInitContextDeadline  = errors.New("deadline context")
	ErrInitConstructorPanic = errors.New("panic in constructor")

	ErrRunContextDeadline = errors.New("running application stopped dua deadline context")
	ErrRunPanic           = errors.New("running application stopped dua panic")
	ErrRunService         = errors.New("running application stopped dua service error")

	ErrTerminateTimeout = errors.New("terminate attempt failed due to timeout")
)
