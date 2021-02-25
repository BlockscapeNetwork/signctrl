package types

import (
	"errors"
	"io/ioutil"
	"log"
)

var (
	// ErrAlreadyStarted is returned if the service has already been started.
	ErrAlreadyStarted = errors.New("service is already started")

	// ErrAlreadyStopped is returned if the service has already been stopped.
	ErrAlreadyStopped = errors.New("service is already stopped")
)

// Service defines a servise that can be started and stopped.
type Service interface {
	// Start the service.
	// An error is returned if the service is already started.
	// If OnStart() returns an error, it will also be returned by Start().
	Start() error
	OnStart() error

	// Stop the service.
	// An error is returned if the service is already stopped.
	// If OnStop() returns an error, it will also be returned by Stop().
	Stop() error
	OnStop()

	// Return true if the service is running, and false if not.
	IsRunning() bool

	// Returns a channel which is closed once the service is stopped.
	Quit() <-chan struct{}

	// Returns a string representation of the service.
	String() string
}

/*
BaseService is a classical inheritance-style service declarations. Services can be
started, and stopped.
Users can override the OnStart/OnStop methods. In the absence of errors, these methods
are guaranteed to be called at most once. If OnStart returns an error, service won't
be marked as started, so the user can call Start again.
The caller must ensure that Start and Stop are not called concurrently.
It is ok to call Stop without calling Start first.

Typical usage:
	type FooService struct {
		BaseService
		// private fields
	}
	func NewFooService() *FooService {
		fs := &FooService{
			// init
		}
		fs.BaseService = *NewBaseService(log, "FooService", fs)
		return fs
	}
	func (fs *FooService) OnStart() error {
		fs.BaseService.OnStart() // Always call the overridden method.
		// initialize private fields
		// start subroutines, etc.
	}
	func (fs *FooService) OnStop() error {
		fs.BaseService.OnStop() // Always call the overridden method.
		// close/destroy private fields
		// stop subroutines, etc.
	}
*/
type BaseService struct {
	Logger  *log.Logger
	name    string
	running bool
	quit    chan struct{}

	// The "subclass" of BaseService
	impl Service
}

// NewBaseService creates a new instance of BaseService.
func NewBaseService(logger *log.Logger, name string, impl Service) *BaseService {
	if logger == nil {
		logger = log.New(ioutil.Discard, "", 0)
	}

	return &BaseService{
		Logger:  logger,
		name:    name,
		running: false,
		quit:    make(chan struct{}),
		impl:    impl,
	}
}

// Start starts a service. An error is returned if the service is already running.
// Implements the Service interface.
func (bs *BaseService) Start() error {
	if bs.running {
		return ErrAlreadyStarted
	}

	bs.Logger.Printf("[DEBUG] signctrl: Starting %v service", bs.name)
	bs.running = true
	if err := bs.impl.OnStart(); err != nil {
		return err
	}

	return nil
}

// OnStart does nothing. This way, users don't need to call BaseService.OnStart().
// Implements the Service interface.
func (bs *BaseService) OnStart() {}

// Stop stops a service and closes the quit channel. An error is returned if the
// service is already stopped.
// Implements the Service interface.
func (bs *BaseService) Stop() error {
	if !bs.running {
		return ErrAlreadyStopped
	}

	bs.Logger.Printf("[DEBUG] signctrl: Stopping %v service", bs.name)
	bs.running = false
	close(bs.quit)
	bs.impl.OnStop()

	return nil
}

// OnStop does nothing. This way, users don't need to call BaseService.OnStop().
// Implements the Service interface.
func (bs *BaseService) OnStop() {}

// IsRunning returns true or false, depending on whether the service is running
// or not.
// Implements the Service interface.
func (bs *BaseService) IsRunning() bool {
	return bs.running
}

// Wait blocks until the service is stopped.
// Implements the Service interface.
func (bs *BaseService) Wait() {
	<-bs.quit
}

// Quit returns a quit channel.
// Implements the Service interface.
func (bs *BaseService) Quit() <-chan struct{} {
	return bs.quit
}

// String returns a string representation of the service.
// Implements the Service interface.
func (bs *BaseService) String() string {
	return bs.name
}
