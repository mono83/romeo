package server

import (
	"os"
	"sync"
	"time"

	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"github.com/mono83/xray/args/env"

	"github.com/mono83/romeo"

	"github.com/mono83/romeo/services"
	"github.com/mono83/romeo/sys"
)

// Server is general all-case services server
type Server struct {
	lock     sync.Mutex
	wait     chan bool
	services services.Container
}

// Register registers one or more services on server
func (s *Server) Register(list ...romeo.Service) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.services == nil {
		s.services = services.Container([]romeo.Service{})
	}

	return s.services.Register(list...)
}

// Join function waits until server stops all services
func (s *Server) Join() {
	s.lock.Lock()
	c := s.wait
	s.lock.Unlock()

	if c != nil {
		<-c
		time.Sleep(50 * time.Millisecond)
	}
}

// Start starts server
func (s *Server) Start(ray xray.Ray) (err error) {
	ray = xray.OrRoot(ray).WithLogger("romeo-server")

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.wait != nil {
		// Already running
		return nil
	}

	if s.services != nil {
		err = s.services.Start(ray)
		if err != nil {
			return
		}
	}

	// Building stop channel
	s.wait = make(chan bool, 1)

	// Starting signal handler
	sys.OnShutdown(func(sig os.Signal) {
		ray.Warning("Received signal :name", args.Name(sig.String()))
		s.Stop(ray.Fork())
	})
	ray.Info("System signals dispatcher started")
	ray.Info("Start sequence done. Running on PID :pid at :hostname", env.PID, env.HostName)
	return
}

// Stop stops server
func (s *Server) Stop(ray xray.Ray) (err error) {
	ray = xray.OrRoot(ray).WithLogger("romeo-server")

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.services != nil {
		err = s.services.Stop(ray)
	}

	s.wait <- true
	s.wait = nil

	return
}
