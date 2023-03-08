package application

import "time"

type Config struct {
	InitialisationTimeout time.Duration `json:"initialisation_timeout"`
	TerminationTimeout    time.Duration `json:"termination_timeout"`
	EnableDebugStack      bool          `json:"enable_debug_stack"`
}
