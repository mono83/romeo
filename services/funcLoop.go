package services

import (
	"sync"
	"time"

	"github.com/mono83/romeo"
	"github.com/mono83/xray"
)

// FuncLoop builds service around target function.
// This function will be invoked at regular basis within tick interval
func FuncLoop(name string, tick time.Duration, target func(xray.Ray)) romeo.Starter {
	return &loop{
		nameHolder: nameHolder(name),
		tick:       tick,
		target:     target,
	}
}

type loop struct {
	nameHolder
	m      sync.Mutex
	tick   time.Duration
	target func(xray.Ray)
	ticker *time.Ticker
}

func (l *loop) Start(r xray.Ray) error {
	l.m.Lock()
	defer l.m.Unlock()

	if l.ticker != nil {
		return romeo.ErrAlreadyRunning{Service: l}
	}

	ticker := time.NewTicker(l.tick)

	go func() {
		for range ticker.C {
			l.target(r.Fork())
		}
	}()
	l.ticker = ticker
	return nil
}

func (l *loop) Stop(r xray.Ray) error {
	l.m.Lock()
	defer l.m.Unlock()

	if l.ticker != nil {
		l.ticker.Stop()
		l.ticker = nil
	}
	return nil
}
