package romeo

// Service is basic interface for services without special capabilities
type Service interface {
	// GetName returns service name
	GetName() string
}
