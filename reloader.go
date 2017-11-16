package romeo

import "time"
import "github.com/mono83/xray"

// Reloader is interface for services, capable to reload
type Reloader interface {
	Service

	// GetReloadInterval return interval, this service should be reloaded
	GetReloadInterval() time.Duration

	// Reload refreshes contents of service
	Reload(xray.Ray) error
}
