package connectivity

import (
	logger "g_matching_engine/pkg/mlog"

	"go.uber.org/zap"
)

// State indicates the state of connectivity.
// It can be the state of a ClientConn or SubConn.
type State int

const (
	// Idle indicates the ClientConn is idle.
	Idle State = iota
	// Connecting indicates the ClientConn is connecting.
	Connecting
	// Ready indicates the ClientConn is ready for work.
	Ready
	// TransientFailure indicates the ClientConn has seen a failure but expects to recover.
	TransientFailure
	// Shutdown indicates the ClientConn has started shutting down.
	Shutdown
)

func (s State) String() string {
	switch s {
	case Idle:
		return "IDLE"
	case Connecting:
		return "CONNECTING"
	case Ready:
		return "READY"
	case TransientFailure:
		return "TRANSIENT_FAILURE"
	case Shutdown:
		return "SHUTDOWN"
	default:
		logger.Logger.Error("unknown connectivity state: ", zap.Any("", s))
		return "Invalid-State"
	}
}
