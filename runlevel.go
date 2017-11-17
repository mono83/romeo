package romeo

import "strconv"

// RunLevel represents service startup priority
// Services with lower priority starts first
// byte is an alias fo uint8, so values must be in range [0, 255]
type RunLevel byte

// List of runlevel constants
// Services with lower run levels invokes first
const (
	RunLevelFirst          RunLevel = 0
	RunLevelDB                      = 100 // i.e. MySQL connection establishment
	RunLevelAfterDB                 = 110
	RunLevelReloaders               = 120 // i.e. load all dictionaries into memory from files/database
	RunLevelAfterReloaders          = 130
	RunLevelAPIServer               = 140 // i.e. start RabbitMQ server but dont start queue listeners
	RunLevelBeforeMain              = 150
	RunLevelMain                    = 160 // i.e. main logic, enpoints
	RunLevelLast                    = 255
)

// WithRunLevel describes structures (and services) that has run priorities
type WithRunLevel interface {
	GetRunLevel() RunLevel
}

func (r RunLevel) String() string {
	prefix := "unknown"
	switch r {
	case RunLevelFirst:
		prefix = "First"
	case RunLevelDB:
		prefix = "Database"
	case RunLevelAfterDB:
		prefix = "After database"
	case RunLevelReloaders:
		prefix = "Reloaders"
	case RunLevelAfterReloaders:
		prefix = "After reloaders"
	case RunLevelAPIServer:
		prefix = "API server"
	case RunLevelBeforeMain:
		prefix = "Before main"
	case RunLevelMain:
		prefix = "Main"
	case RunLevelLast:
		prefix = "Last"
	}

	return prefix + " (" + strconv.Itoa(int(r)) + ")"
}

// RunLevelForService calculates run level for service
func RunLevelForService(s Service) RunLevel {
	if s == nil {
		return RunLevelFirst
	}
	if r, ok := s.(WithRunLevel); ok {
		return r.GetRunLevel()
	}

	return RunLevelMain
}
