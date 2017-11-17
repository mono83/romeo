package services

import (
	"time"

	"github.com/mono83/romeo"
	"github.com/mono83/xray"
)

// FuncLoop builds service around target function.
// This function will be invoked at regular basis within tick interval
func FuncLoop(name string, tick time.Duration, target func(xray.Ray) error) romeo.Service {
	return &Reloader{Reloadable: loopFuncAdapter{name: name, target: target, tick: tick}}
}

type loopFuncAdapter struct {
	name   string
	target func(xray.Ray) error
	tick   time.Duration
}

func (l loopFuncAdapter) GetName() string                  { return l.name }
func (l loopFuncAdapter) GetReloadInterval() time.Duration { return l.tick }
func (l loopFuncAdapter) Reload(r xray.Ray) error          { return l.target(r) }
