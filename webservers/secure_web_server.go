package webservers

import (
	"net/http"

	"github.com/off-sync/platform-proxy-app/interfaces"
)

// SecureWebServer implements the Secure Web Server interface, extending the
// WebServer with certificate management.
type SecureWebServer struct {
	*routeManager
	*certificateManager

	server *http.Server
}

// NewSecureWebServer creates a new secure web server listening on the provided
// address.
func NewSecureWebServer(log interfaces.Logger, addr string) *SecureWebServer {
	s := &SecureWebServer{
		routeManager:       newRouteManager(log),
		certificateManager: newCertificateManager(),
		server: &http.Server{
			Addr: addr,
		},
	}

	s.server.Handler = s

	return s
}
