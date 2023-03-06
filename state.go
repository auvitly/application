package application

type State int

const (
	StateInit State = iota
	StateReady
	StateRunning
	StateShutdown
	StateOff
)

func (s State) String() string {
	switch s {
	case StateInit:
		return "state_init"
	case StateReady:
		return "state_ready"
	case StateRunning:
		return "state_running"
	case StateShutdown:
		return "state_shutdown"
	case StateOff:
		return "state_off"
	default:
		return "state_error"
	}
}
