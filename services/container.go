package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/mono83/romeo"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"github.com/mono83/xray/args/env"
)

// Container is special service, that is just slice of other services
type Container []romeo.Service

// GetName returns service's name, for this case it would be "container"
func (Container) GetName() string { return "container" }

// Register registers one or more services within container
func (c *Container) Register(services ...romeo.Service) error {
	*c = append(*c, services...)
	fmt.Println(c)
	return nil
}

// Start starts services container
func (c Container) Start(ray xray.Ray) error {
	ray = ray.WithLogger("container")

	ray.Debug("Starting services container with :count services in total", args.Count(len(c)))

	allBefore := time.Now()
	for _, group := range romeo.GroupByRunLevel(c, false) {
		rl := romeo.RunLevelForService(group[0])
		wg := sync.WaitGroup{}
		wg.Add(len(group))
		ray.Debug("Starting :count services on run level :level", args.Count(len(group)), args.String{N: "level", V: rl.String()})

		var resultingError error
		for _, service := range group {
			go func(s romeo.Service) {
				serviceLog := ray.With(args.Name(romeo.NameForService(s)))
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
	return nil
}

// Stop stops services container
func (c Container) Stop(ray xray.Ray) error {
	ray = ray.WithLogger("container")

	ray.Debug("Stopping services container")

	allBefore := time.Now()
	for _, group := range romeo.GroupByRunLevel(c, true) {
		rl := romeo.RunLevelForService(group[0])
		wg := sync.WaitGroup{}
		wg.Add(len(group))

		for _, service := range group {
			go func(s romeo.Service) {
				serviceLog := ray.With(args.Name(romeo.NameForService(s)))
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
	return nil
}

// Size returns amount of services within container
func (c Container) Size() int {
	return len(c)
}
