package romeo

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/mono83/romeo/sig"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
)

// SimpleServer is simple services server implementation
type SimpleServer struct {
	m         sync.Mutex
	wait      chan bool
	starters  map[RunLevel][]Starter
	reloaders []*reloader
}

// registerStarter is utility method, used to register services, able to start
func (s *SimpleServer) registerStarter(service Starter) {
	if s.starters == nil {
		s.starters = map[RunLevel][]Starter{}
	}

	s.starters[service.GetRunLevel()] = append(s.starters[service.GetRunLevel()], service)
}

// Register registers one or more services on server
func (s *SimpleServer) Register(services ...Service) error {
	s.m.Lock()
	defer s.m.Unlock()

	if len(services) > 0 {
		for _, service := range services {
			if service == nil {
				continue
			}
			if x, ok := service.(Starter); ok {
				s.registerStarter(x)
			}
			if x, ok := service.(Reloader); ok {
				rc := &reloader{Reloader: x}
				s.reloaders = append(s.reloaders, rc)
				s.registerStarter(rc)
			}
		}
	}
	return nil
}

// Start starts all services on server
func (s *SimpleServer) Start(ray xray.Ray) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.wait != nil {
		// Already running
		return nil
	}

	if ray == nil {
		ray = xray.BOOT
	}
	ray = ray.WithLogger("simple-server")

	ray.Debug("Starting services container service")

	// Starting all services
	var i int
	allBefore := time.Now()
	for i = 0; i <= 255; i++ {
		rl := RunLevel(i)
		services, ok := s.starters[rl]
		if !ok || len(services) == 0 {
			// No services on runlevel, skip
			continue
		}

		// Running all services in parallel
		wg := sync.WaitGroup{}
		wg.Add(len(services))
		ray.Debug("Starting :count services on run level :level", args.Count(len(services)), args.String{N: "level", V: rl.String()})
		var resultingError error
		for _, service := range services {
			go func(s Starter) {
				serviceLog := ray.With(args.Name(s.GetName()))
				serviceLog.Debug("Starting service :name")
				before := time.Now()
				if err := s.Start(serviceLog); err != nil {
					serviceLog.Error("Service :name start failed with :err", args.Error{Err: err})
					if resultingError == nil {
						resultingError = err
					}
				} else {
					serviceLog.Info("Service :name started in :delta", args.Delta(time.Now().Sub(before)))
				}
				wg.Done()
			}(service)
		}
		wg.Wait()

		if resultingError != nil {
			return resultingError
		}
	}

	ray.Info("Startup sequence done in :delta without errors", args.Delta(time.Now().Sub(allBefore)))

	// Building stop channel
	s.wait = make(chan bool, 1)

	// Starting signal handler
	sig.OnShutdown(func(sig os.Signal) {
		ray.Warning("Received signal :name", args.Name(sig.String()))
		s.Stop(ray.Fork())
	})
	ray.Info("System signals dispatcher started")

	return nil
}

// Stop stops all services on server
func (s *SimpleServer) Stop(ray xray.Ray) error {
	s.m.Lock()
	defer s.m.Unlock()

	if ray == nil {
		ray = xray.BOOT
	}
	ray = ray.WithLogger("simple-server")

	var i int
	allBefore := time.Now()
	for i = 255; i >= 0; i-- {
		rl := RunLevel(i)
		services, ok := s.starters[rl]
		if !ok || len(services) == 0 {
			// No services on runlevel, skip
			continue
		}

		// Running all services in parallel
		wg := sync.WaitGroup{}
		wg.Add(len(services))
		for _, service := range services {
			st, ok := service.(Stopper)
			if !ok {
				continue
			}
			go func(s Stopper) {
				serviceLog := ray.With(args.Name(s.GetName()))
				serviceLog.Debug("Stopping service :name on :level", args.String{N: "level", V: rl.String()})
				if err := s.Stop(serviceLog); err != nil {
					serviceLog.Error("Service :name shutdown failed with :err", args.Error{Err: err})
				}
				wg.Done()
			}(st)
		}
		wg.Wait()
	}

	ray.Info("Shutdown sequence done in :delta", args.Delta(time.Now().Sub(allBefore)))

	s.wait <- true
	s.wait = nil

	return nil
}

// Join function waits until server stops all services
func (s *SimpleServer) Join() {
	s.m.Lock()
	c := s.wait
	s.m.Unlock()

	if c != nil {
		<-c
		time.Sleep(50 * time.Millisecond)
	}
}

type reloader struct {
	m      sync.Mutex
	ticker *time.Ticker
	Reloader
}

// GetRunLevel is Starter interface implementation
func (*reloader) GetRunLevel() RunLevel { return RunLevelReloaders }

// Stop is Stopper interface implementation
func (r *reloader) Stop(ray xray.Ray) error {
	r.m.Lock()
	defer r.m.Unlock()
	if r.ticker != nil {
		r.ticker.Stop()
		r.ticker = nil
	}
	return nil
}

// Start is Starter interface implementation
func (r *reloader) Start(ray xray.Ray) error {
	r.m.Lock()
	defer r.m.Unlock()
	if r.ticker != nil {
		return ErrAlreadyRunning{Service: r}
	}
	delta := r.GetReloadInterval()
	if delta.Seconds() < 0.1 {
		return errors.New("at least 100ms must be configured for reloader")
	}
	if err := r.Reload(ray); err != nil {
		return err
	}

	// Starting ticker
	r.ticker = time.NewTicker(delta)
	go func() {
		for range r.ticker.C {
			r.Reload(ray.Fork())
		}
	}()

	return nil
}
