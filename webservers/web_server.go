package webservers

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"github.com/gorilla/mux"
	"github.com/off-sync/platform-proxy-app/interfaces"
)

// Errors.
var (
	ErrMissingRoute   = errors.New("route must not be nil")
	ErrMissingHandler = errors.New("handler must not be nil")
)

// WebServer implements the Web Server interface. It uses the Gorilla Mux as
// its router.
type WebServer struct {
	log interfaces.Logger

	routesLock sync.Mutex
	routes     webServerRoutes

	routerLock sync.RWMutex
	router     *mux.Router
	routerHash string

	server *http.Server
}

type webServerRoutes map[string]*webServerRoute

func (r webServerRoutes) hash() string {
	buf := &bytes.Buffer{}

	var keys []string
	for key := range r {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		route := r[key]
		fmt.Fprintf(buf, "%s:%p", key, route.handler)
	}

	h := sha256.Sum256(buf.Bytes())
	return string(h[:])
}

type webServerRoute struct {
	url     *url.URL
	handler http.Handler
}

func newWebServerRoute(url *url.URL, handler http.Handler) *webServerRoute {
	return &webServerRoute{
		url:     url,
		handler: handler,
	}
}

func (r *webServerRoute) key() string {
	return fmt.Sprintf("%s|%s", r.url.Hostname(), r.url.Path)
}

// NewWebServer creates a new Web Server listening on the provided address.
func NewWebServer(log interfaces.Logger, addr string) *WebServer {
	router := mux.NewRouter()
	router.StrictSlash(true)

	webServer := &WebServer{
		log:    log,
		routes: make(map[string]*webServerRoute),
		router: router,
		server: &http.Server{
			Addr: addr,
		},
	}

	webServer.server.Handler = webServer

	go func() {
		if err := webServer.server.ListenAndServe(); err != nil {
			panic(err) // TODO
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

func (s *WebServer) reconfigureRoutes() {
	// s.routes is already locked when reconfigureRoutes is called
	hash := s.routes.hash()
	if hash == s.routerHash {
		// routes haven't been changed
		return
	}

	// configure a new router
	router := mux.NewRouter()

	for _, route := range s.routes {
		router.
			Host(route.url.Hostname()).
			PathPrefix(route.url.Path).
			Handler(route.handler)
	}

	// update the router
	s.routerLock.Lock()
	s.router = router
	s.routerHash = hash
	s.routerLock.Unlock()

	return
}

// UpsertRoute adds a route to the web server, forwarding all requests to the
// provided handler. It returns an error if either parameter is nil.
func (s *WebServer) UpsertRoute(route *url.URL, handler http.Handler) error {
	if route == nil {
		return ErrMissingRoute
	}

	if handler == nil {
		return ErrMissingHandler
	}

	s.routesLock.Lock()
	defer s.routesLock.Unlock()

	r := newWebServerRoute(route, handler)

	s.routes[r.key()] = r

	s.reconfigureRoutes()

	return nil
}

// DeleteRoute deletes a route from the web server.
func (s *WebServer) DeleteRoute(route *url.URL) {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()

	r := newWebServerRoute(route, nil)

	delete(s.routes, r.key())

	s.reconfigureRoutes()
}
