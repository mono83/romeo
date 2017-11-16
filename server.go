package romeo

import (
	"github.com/mono83/xray"
)

// Server is component, able to start, stop and reload other components
type Server interface {
	// Register registers one or more services on server
	Register(...Service) error
	// Start starts whole server and it's services
	Start(xray.Ray) error
	// Stop stops whole server and it's services
	Stop(xray.Ray) error
}
