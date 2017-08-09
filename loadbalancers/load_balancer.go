package loadbalancers

import (
	"fmt"
	"net/http"
	"net/url"
)

// LoadBalancer implements the Load Balancer interface.
type LoadBalancer struct {
}

// UpsertService sets the urls for a service. It returns an http.Handler
// that can be used to make requests to for this service.
func (l *LoadBalancer) UpsertService(name string, urls ...*url.URL) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintf(w, "service: %s\nurls: %v", name, urls)
	}), nil
}

// DeleteService deletes the service from the load balancer.
func (l *LoadBalancer) DeleteService(name string) {
	// do nothing
}
