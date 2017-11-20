package romeo

import (
	"fmt"

	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
)

// Named describes structures (and services) that has names
type Named interface {
	GetName() string
}

// NameForService returns service name
func NameForService(s interface{}) string {
	if s == nil {
		return ""
	}
	if n, ok := s.(Named); ok {
		return n.GetName()
	}

	return fmt.Sprintf("Service of type %T", s)
}

// ArgForService returns xray.Arg, containing key "name"
// and service name itself
func ArgForService(s Service) xray.Arg {
	return args.Name(NameForService(s))
}
