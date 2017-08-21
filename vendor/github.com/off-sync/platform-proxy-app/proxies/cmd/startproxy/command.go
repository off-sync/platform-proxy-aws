package startproxy

import (
	"context"
	"errors"
	"sync"

	"github.com/off-sync/platform-proxy-app/interfaces"
)

// Errors
var (
	ErrFrontendRepositoryMissing = errors.New("frontend repository missing")
	ErrServiceRepositoryMissing  = errors.New("service repository missing")
	ErrWebServerMissing          = errors.New("web server missing")
	ErrSecureWebServerMissing    = errors.New("secure web server missing")
	ErrLoadBalancerMissing       = errors.New("load balancer missing")
	ErrInvalidPollingDuration    = errors.New("invalid polling duration, must greater than or equal to 0")
)

// Command models the Start Proxy Command which can be used to start one of the
// platform proxies.
type Command struct {
	serviceRepository  interfaces.ServiceRepository
	frontendRepository interfaces.FrontendRepository
	logger             interfaces.Logger
}

// NewCommand creates a new Start Proxy Command using the provided frontend
// and service repositories.
func NewCommand(
	serviceRepository interfaces.ServiceRepository,
	frontendRepository interfaces.FrontendRepository,
	logger interfaces.Logger) (*Command, error) {
	if serviceRepository == nil {
		return nil, ErrServiceRepositoryMissing
	}

	if frontendRepository == nil {
		return nil, ErrFrontendRepositoryMissing
	}

	return &Command{
		serviceRepository:  serviceRepository,
		frontendRepository: frontendRepository,
		logger:             logger,
	}, nil
}

// Execute runs the Start Proxy Command by configuring the required listeners.
func (c *Command) Execute(model *Model) error {
	if model.WebServer == nil {
		return ErrWebServerMissing
	}

	if model.SecureWebServer == nil {
		return ErrSecureWebServerMissing
	}

	if model.LoadBalancer == nil {
		return ErrLoadBalancerMissing
	}

	if model.PollingDuration < 1 {
		return ErrInvalidPollingDuration
	}

	if model.Ctx == nil {
		model.Ctx = context.Background()
	}

	if model.WaitGroup == nil {
		model.WaitGroup = &sync.WaitGroup{}
	}

	proxy := newProxy(
		model.Ctx,
		model.WaitGroup,
		c.logger,
		c.serviceRepository,
		c.frontendRepository,
		model.PollingDuration,
		model.WebServer,
		model.SecureWebServer,
		model.LoadBalancer)

	go proxy.run()

	return nil
}
