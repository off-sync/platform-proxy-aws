package loadbalancers

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

// LoadBalancer implements the Load Balancer interface.
type LoadBalancer struct {
}

// UpsertService sets the urls for a service. It returns an http.Handler
// that can be used to make requests to for this service.
func (l *LoadBalancer) UpsertService(name string, urls ...*url.URL) (http.Handler, error) {
	fwd, err := forward.New()
	if err != nil {
		return nil, err
	}

	lb, err := roundrobin.New(fwd)
	if err != nil {
		return nil, err
	}

	for _, u := range urls {
		addrs, err := net.LookupHost(u.Hostname())
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			u2 := &url.URL{}
			*u2 = *u
			u2.Host = fmt.Sprintf("%s:%s", addr, u.Port())

			lb.UpsertServer(u2)
		}
	}

	return lb, nil
}

// DeleteService deletes the service from the load balancer.
func (l *LoadBalancer) DeleteService(name string) {
	// do nothing
}
