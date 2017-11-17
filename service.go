package romeo

import (
	"github.com/mono83/xray"
)

// Service is basic interface for services without special capabilities
type Service interface {
	// Start starts service
	Start(xray.Ray) error
	// Stop stops service
	Stop(xray.Ray) error
}

// ErrAlreadyRunning is an error, emitted on attemp to start running service
type ErrAlreadyRunning struct {
	Service Service
}

func (e ErrAlreadyRunning) Error() string {
	if e.Service != nil {
		if n, ok := e.Service.(Named); ok {
			return "service " + n.GetName() + " already running"
		}
	}

	return "service already running"
}
