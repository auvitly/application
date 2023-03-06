package application

import "time"

type Config struct {
	InitialisationTimeout time.Duration
	TerminationTimeout    time.Duration
	EnableDebugStack      bool
}
