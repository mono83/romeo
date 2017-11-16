package services

import (
	"sync"

	"github.com/mono83/romeo"
	"github.com/mono83/xray"
)

// FuncStopable builds service around stopable function
// Stop signal will be emitted throught bool chan
func FuncStopable(name string, target func(<-chan bool)) romeo.Service {
	return &stoppable{
		nameHolder: nameHolder(name),
		target:     target,
	}
}

type stoppable struct {
	nameHolder
	m      sync.Mutex
	ch     chan bool
	target func(<-chan bool)
}

func (s *stoppable) Start(r xray.Ray) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.ch != nil {
		return romeo.ErrAlreadyRunning{Service: s}
	}

	s.ch = make(chan bool)
	go s.target(s.ch)
	return nil
}

func (s *stoppable) Stop(r xray.Ray) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.ch != nil {
		s.ch <- true
		s.ch = nil
	}
	return nil
}
