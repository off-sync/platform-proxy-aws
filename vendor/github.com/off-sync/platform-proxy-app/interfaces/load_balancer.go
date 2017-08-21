package interfaces

import (
	"net/http"
	"net/url"
)

// LoadBalancer defines the interface for a load balancer that distributes
// requests for a service to different endpoints defined as URLs.
type LoadBalancer interface {
	// UpsertService sets the urls for a service. It returns an http.Handler
	// that can be used to make requests to for this service.
	UpsertService(name string, urls ...*url.URL) (http.Handler, error)

	// DeleteService deletes the service from the load balancer.
	DeleteService(name string)
}
