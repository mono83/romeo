package services

import (
	"sync"
	"time"

	"github.com/mono83/romeo"
	"github.com/mono83/xray"
)

// Reloadable describes structures, that can be wrapped into Reloader
// to periodically refresh data
type Reloadable interface {
	GetReloadInterval() time.Duration
	Reload(xray.Ray) error
}

// Reloader is service, that periofically invokes provided func
type Reloader struct {
	Reloadable
	nameHolder
	m      sync.Mutex
	ticker *time.Ticker
}

// GetRunLevel return run level for service
func (*Reloader) GetRunLevel() romeo.RunLevel { return romeo.RunLevelReloaders }

// Start starts service
func (r *Reloader) Start(ray xray.Ray) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.ticker != nil {
		return romeo.ErrAlreadyRunning{Service: r}
	}

	ray.Trace("reloader: initial start for :name", romeo.ArgForService(r))
	if err := r.Reload(ray); err != nil {
		return err
	}

	ticker := time.NewTicker(r.GetReloadInterval())

	go func() {
		for range ticker.C {
			forked := ray.Fork()
			ray.Trace("reloader: reloading :name", romeo.ArgForService(r))
			r.Reload(forked)
		}
	}()
	r.ticker = ticker
	return nil
}

// Stop stops service
func (r *Reloader) Stop(ray xray.Ray) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.ticker != nil {
		ray.Trace("reloader: stopping ticker for :name", romeo.ArgForService(r))
		r.ticker.Stop()
		r.ticker = nil
	}
	return nil
}
