package types

import "syscall"

type OperationResult string

const (
	ResultSuccess OperationResult = "result_success"
	ResultError   OperationResult = "result_error"
)

const (
	SIGPANIC = syscall.Signal(0x0)
)
