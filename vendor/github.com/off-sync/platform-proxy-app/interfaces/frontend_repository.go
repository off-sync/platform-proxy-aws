package interfaces

import (
	"errors"

	"github.com/off-sync/platform-proxy-domain/frontends"
)

// Errors
var (
	ErrUnknownFrontend = errors.New("unknown frontend")
)

// FrontendRepository is a repository for frontends.
type FrontendRepository interface {
	// ListFrontends returns all frontend names contained in this repository.
	ListFrontends() ([]string, error)

	// DescribeFrontend returns the frontend with the specified name. If no
	// frontend	exists with that name an ErrUnknownFrontend is returned.
	DescribeFrontend(name string) (*frontends.Frontend, error)
}
