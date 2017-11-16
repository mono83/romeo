package romeo

import (
	"github.com/mono83/xray"
)

// Stopper is interface for services, capable to start and stop
type Stopper interface {
	Starter

	// Stop stops service
	Stop(xray.Ray) error
}
