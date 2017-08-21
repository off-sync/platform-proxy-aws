package interfaces

import (
	"errors"

	"github.com/off-sync/platform-proxy-domain/services"
)

// Errors
var (
	ErrUnknownService = errors.New("unknown service")
)

// ServiceRepository is a repository for services.
type ServiceRepository interface {
	// ListServices returns all service names contained in this repository.
	ListServices() ([]string, error)

	// DescribeService returns the service with the specified name. If no service
	// exists with that name an ErrUnknownService is returned.
	DescribeService(name string) (*services.Service, error)
}
