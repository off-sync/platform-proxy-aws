package webservers

import (
	"fmt"
	"net/http"
	"net/url"
)

type route struct {
	url     *url.URL
	handler http.Handler
}

func newRoute(url *url.URL, handler http.Handler) *route {
	return &route{
		url:     url,
		handler: handler,
	}
}

func (r *route) key() string {
	return fmt.Sprintf("%s|%s", r.url.Hostname(), r.url.Path)
}
