package interfaces

import (
	"net/http"
	"net/url"

	"github.com/off-sync/platform-proxy-domain/frontends"
)

// WebServer defines an interface provides methods to upsert and delete routes
// to a handler.
type WebServer interface {
	// UpsertRoute adds a route to the web server, forwarding all requests to the
	// provided handler. It returns an error if either parameter is nil.
	UpsertRoute(route *url.URL, handler http.Handler) error

	// DeleteRoute deletes a route from the web server.
	DeleteRoute(route *url.URL)
}

// SecureWebServer extends WebServer with the option to provide certificates
// for domain names.
type SecureWebServer interface {
	WebServer

	// UpsertCertificate sets the certificate for the provided domain name.
	UpsertCertificate(domainName string, cert *frontends.Certificate) error
}
