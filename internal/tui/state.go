package tui

type state int

const (
	stateHealth state = iota
	stateAutoFix
	stateList
	stateSearch
	stateConfirm
	stateRunning
	stateOutput
	stateError
	stateAdmin
	stateLogs
	stateMonitor
)

func (s state) String() string {
	switch s {
	case stateHealth:
		return "health"
	case stateAutoFix:
		return "autofix"
	case stateList:
		return "list"
	case stateSearch:
		return "search"
	case stateConfirm:
		return "confirm"
	case stateRunning:
		return "running"
	case stateOutput:
		return "output"
	case stateError:
		return "error"
	case stateAdmin:
		return "admin"
	case stateLogs:
		return "logs"
	case stateMonitor:
		return "monitor"
	default:
		return "unknown"
	}
}
