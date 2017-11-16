package romeo

import (
	"github.com/mono83/xray"
)

// Starter is interface for services, capable to start
type Starter interface {
	Service

	// GetRunLevel returns runlevel for service
	GetRunLevel() RunLevel

	// Start starts service
	// Will return error (ErrAlreadyRunning is most cases) if service already running
	Start(xray.Ray) error
}

// ErrAlreadyRunning is an error, emitted on attemp to start running service
type ErrAlreadyRunning struct {
	Service Service
}

func (e ErrAlreadyRunning) Error() string {
	if e.Service == nil {
		return "service already running"
	}
	return "service " + e.Service.GetName() + " already running"
}
