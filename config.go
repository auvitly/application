package application

import "time"

// Config - configuration application struct.
type Config struct {
	InitialisationTimeout time.Duration `json:"initialisation_timeout"`
	TerminationTimeout    time.Duration `json:"termination_timeout"`
	EnableDebugStack      bool          `json:"enable_debug_stack"`
}
