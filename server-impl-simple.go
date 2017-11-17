package romeo

import (
	"os"
	"sync"
	"time"

	"github.com/mono83/romeo/env"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
)

// SimpleServer is simple services server implementation
type SimpleServer struct {
	m        sync.Mutex
	wait     chan bool
	services map[RunLevel][]Service
}

// register is utility method, used to register services, able to start
func (s *SimpleServer) register(service Service) {
	if s.services == nil {
		s.services = map[RunLevel][]Service{}
	}

	rl := RunLevelForService(service)

	s.services[rl] = append(s.services[rl], service)
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
			s.register(service)
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
		services, ok := s.services[rl]
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
			go func(s Service) {
				serviceLog := ray.With(args.Name(NameForService(s)))
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
	env.OnShutdown(func(sig os.Signal) {
		ray.Warning("Received signal :name", args.Name(sig.String()))
		s.Stop(ray.Fork())
	})
	ray.Info("System signals dispatcher started")
	ray.Info("Start sequence done. Running on PID :pid at :hostname", env.PID, env.HostName)

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
		services, ok := s.services[rl]
		if !ok || len(services) == 0 {
			// No services on runlevel, skip
			continue
		}

		// Running all services in parallel
		wg := sync.WaitGroup{}
		wg.Add(len(services))
		for _, service := range services {
			go func(s Service) {
				serviceLog := ray.With(args.Name(NameForService(s)))
				serviceLog.Debug("Stopping service :name on :level", args.String{N: "level", V: rl.String()})
				if err := s.Stop(serviceLog); err != nil {
					serviceLog.Error("Service :name shutdown failed with :err", args.Error{Err: err})
				}
				wg.Done()
			}(service)
		}
		wg.Wait()
	}

	ray.Info("Shutdown sequence done in :delta. PID :pid", args.Delta(time.Now().Sub(allBefore)), env.PID)

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
