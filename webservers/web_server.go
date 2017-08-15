package webservers

import (
	"errors"
	"net/http"

	"github.com/off-sync/platform-proxy-app/interfaces"
)

// Errors.
var (
	ErrServerStartupTimout = errors.New("server startup took too long")
)

// WebServer implements the Web Server interface. It uses the Gorilla Mux as
// its router.
type WebServer struct {
	*routeManager

	server *http.Server
}

// NewWebServer creates a new Web Server listening on the provided address.
func NewWebServer(log interfaces.Logger, addr string) *WebServer {
	s := &WebServer{
		routeManager: newRouteManager(log),
		server: &http.Server{
			Addr: addr,
		},
	}

	s.server.Handler = s

	// start the http server
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.WithError(err).Error("listening and serving")
		}
	}()

	return s
}
