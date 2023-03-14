package application

type state int

const (
	stateInit state = iota
	stateReady
	stateRunning
	stateShutdown
	stateOff
)

func (s state) String() string {
	switch s {
	case stateInit:
		return "state_init"
	case stateReady:
		return "state_ready"
	case stateRunning:
		return "state_running"
	case stateShutdown:
		return "state_shutdown"
	case stateOff:
		return "state_off"
	default:
		return "state_error"
	}
}
