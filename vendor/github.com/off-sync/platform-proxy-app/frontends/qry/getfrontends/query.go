package getfrontends

import (
	"errors"

	"github.com/off-sync/platform-proxy-app/interfaces"
	"github.com/off-sync/platform-proxy-domain/frontends"
)

// Errors
var (
	ErrMissingFrontendRepository = errors.New("missing frontends repository")
)

// Query implements the Get Frontends Query. It requires a FrontendRepository.
type Query struct {
	repo interfaces.FrontendRepository
}

// NewQuery creates a new Get Frontends Query
func NewQuery(repo interfaces.FrontendRepository) (*Query, error) {
	if repo == nil {
		return nil, ErrMissingFrontendRepository
	}

	return &Query{
		repo: repo,
	}, nil
}

// Execute performs the Get Frontends Query using the provided model.
func (q *Query) Execute(model *Model) (*Result, error) {
	names, err := q.repo.ListFrontends()
	if err != nil {
		return nil, err
	}

	result := &Result{
		Frontends: make([]*frontends.Frontend, len(names)),
	}

	for i, name := range names {
		result.Frontends[i], err = q.repo.DescribeFrontend(name)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
