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
	m := newRouteManager(log)

	webServer := &WebServer{
		routeManager: m,
		server: &http.Server{
			Addr: addr,
		},
	}

	webServer.server.Handler = webServer

	// start the http server
	go func() {
		if err := webServer.server.ListenAndServe(); err != nil {
			log.WithError(err).Error("listening and serving")
		}
	}()

	return webServer
}

// ServeHTTP process requests using the configured router.
func (s *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log.WithField("host", r.Host).
		WithField("url", r.URL).
		Debug("serving HTTP")

	s.routerLock.RLock()
	router := s.router
	s.routerLock.RUnlock()

	router.ServeHTTP(w, r)
}
