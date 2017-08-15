package webservers

import (
	"errors"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/mux"
	"github.com/off-sync/platform-proxy-app/interfaces"
)

// Route Manager errors.
var (
	ErrMissingRoute   = errors.New("route must not be nil")
	ErrMissingHandler = errors.New("handler must not be nil")
)

type routeManager struct {
	log interfaces.Logger

	routesLock sync.Mutex
	routes     routes

	routerLock sync.RWMutex
	router     *mux.Router
	routerHash string
}

func newRouteManager(log interfaces.Logger) *routeManager {
	router := mux.NewRouter()

	m := &routeManager{
		log:    log,
		routes: make(routes),
		router: router,
	}

	return m
}

func (m *routeManager) reconfigureRoutes() {
	// s.routes is already locked when reconfigureRoutes is called
	hash := m.routes.hash()
	if hash == m.routerHash {
		// routes haven't been changed
		return
	}

	// configure a new router
	router := mux.NewRouter()

	for _, route := range m.routes {
		path := route.url.Path
		if path == "" {
			path = "/"
		}

		router.
			PathPrefix(path).
			Host(route.url.Hostname()).
			Handler(route.handler)
	}

	// update the router
	m.routerLock.Lock()
	defer m.routerLock.Unlock()
	m.router = router
	m.routerHash = hash

	return
}

// UpsertRoute adds a route to the web server, forwarding all requests to the
// provided handler. It returns an error if either parameter is nil.
func (m *routeManager) UpsertRoute(route *url.URL, handler http.Handler) error {
	if route == nil {
		return ErrMissingRoute
	}

	if handler == nil {
		return ErrMissingHandler
	}

	m.routesLock.Lock()
	defer m.routesLock.Unlock()

	r := newRoute(route, handler)

	m.routes[r.key()] = r

	m.reconfigureRoutes()

	return nil
}

// DeleteRoute deletes a route from the web server.
func (m *routeManager) DeleteRoute(route *url.URL) {
	m.routesLock.Lock()
	defer m.routesLock.Unlock()

	r := newRoute(route, nil)

	delete(m.routes, r.key())

	m.reconfigureRoutes()
}

// ServeHTTP process requests using the configured router.
func (m *routeManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.log.WithField("host", r.Host).
		WithField("url", r.URL).
		Debug("serving HTTP")

	m.routerLock.RLock()
	router := m.router
	m.routerLock.RUnlock()

	router.ServeHTTP(w, r)
}
