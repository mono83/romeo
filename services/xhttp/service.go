package xhttp

import (
	"net"
	"net/http"
	"sync"

	"github.com/mono83/xray"
	"github.com/mono83/xray/args"

	"github.com/mono83/romeo"
)

// Service is basic HTTP service, built on top of http.Handler
type Service struct {
	Name    string
	Bind    string
	Handler http.Handler

	lock sync.Mutex
	tcp  net.Listener
	http *http.Server
}

// GetName returns service name
func (s *Service) GetName() string {
	if len(s.Name) == 0 {
		return "HTTP Service on " + s.Bind
	}

	return s.Name
}

// GetRunLevel return service startup priority
func (s *Service) GetRunLevel() romeo.RunLevel { return romeo.RunLevelMain }

// Start starts HTTP service
func (s *Service) Start(ray xray.Ray) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.http != nil {
		return romeo.ErrAlreadyRunning{Service: s}
	}

	ray = ray.WithLogger("http-service")
	ray.Debug("Starting service :name on :addr", args.Name(s.GetName()), args.String{N: "addr", V: s.Bind})

	addr, err := net.ResolveTCPAddr("tcp", s.Bind)
	if err != nil {
		return ray.PassS("Unable to resolve binding address", err)
	}
	ray.Trace("Binding address resolved")

	s.tcp, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return ray.PassS("Unable to bind to address", err)
	}
	ray.Debug("Listening port is up, service is ready")

	s.http = &http.Server{Handler: s.Handler}
	go func() {
		s.http.Serve(s.tcp)
	}()
	return nil
}

// Stop stops HTTP service
func (s *Service) Stop(ray xray.Ray) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.http == nil {
		return nil
	}

	ray = ray.WithLogger("http-service")
	ray.Debug("Starting service :name", args.Name(s.GetName()))

	s.http.Close()
	s.tcp.Close()
	s.http = nil
	s.tcp = nil
	return nil
}
