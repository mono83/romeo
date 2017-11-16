package sig

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// OnShutdown registers handlers, that would be invoked on application start
func OnShutdown(handler func(os.Signal)) {
	OnSignal(handler, os.Interrupt, syscall.SIGTERM)
}

// OnSignal registers handler, that would be invoked on incoming OS signal(s)
func OnSignal(handler func(os.Signal), signals ...os.Signal) {
	if handler == nil || len(signals) == 0 {
		return
	}

	sh := &sigHandler{
		signals: signals,
		target:  handler,
	}
	sh.Start()
}

// sigHandler is helper structure, used to catch system signals and invoke
// corresponding functions
type sigHandler struct {
	m       sync.Mutex
	ch      chan os.Signal
	signals []os.Signal
	target  func(os.Signal)
}

// Start method starts signal handler
func (s *sigHandler) Start() {
	s.m.Lock()
	defer s.m.Unlock()

	if s.ch != nil {
		return
	}

	s.ch = make(chan os.Signal, 1)
	for _, sig := range s.signals {
		signal.Notify(s.ch, sig)
	}
	go func() {
		for in := range s.ch {
			for _, sig := range s.signals {
				if sig == in {
					go s.target(sig)
					break
				}
			}
		}
	}()
}
