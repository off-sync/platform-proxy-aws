package webservers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/off-sync/platform-proxy-app/infra/logging"
	"github.com/stretchr/testify/assert"
)

func setUp(t *testing.T) *WebServer {
	s, err := NewWebServer(logging.NewLogrusLogger(logrus.New()), ":0")
	assert.Nil(t, err)

	return s
}

func TestNewWebServer(t *testing.T) {
	w := setUp(t)
	assert.NotNil(t, w)
}

func TestUpsertRoute(t *testing.T) {
	w := setUp(t)
	handler := newDummyHandler(t, "")

	route, _ := url.Parse("http://localhost/api")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/test", nil)

	w.ServeHTTP(rw, req)
	body, _ := ioutil.ReadAll(rw.Body)

	assert.Equal(t, "[] GET http://localhost/api/test", string(body))
}

func TestUpsertRouteShouldReturnErrorOnMissingRoute(t *testing.T) {
	w := setUp(t)
	handler := newDummyHandler(t, "")

	err := w.UpsertRoute(nil, handler)
	assert.Equal(t, ErrMissingRoute, err)
}

func TestUpsertRouteShouldReturnErrorOnMissingHandler(t *testing.T) {
	w := setUp(t)
	route, _ := url.Parse("http://localhost/api")

	err := w.UpsertRoute(route, nil)
	assert.Equal(t, ErrMissingHandler, err)
}

func TestUpsertRouteShouldAllowUpdating(t *testing.T) {
	w := setUp(t)
	route, _ := url.Parse("http://localhost/api/")

	handler := newDummyHandler(t, "h1")
	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	handler = newDummyHandler(t, "h2")
	err = w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/test", nil)

	w.ServeHTTP(rw, req)
	body, _ := ioutil.ReadAll(rw.Body)

	assert.Equal(t, "[h2] GET http://localhost/api/test", string(body))
}

func TestUpsertRouteShouldNotUpdateRouterIfRoutesUnchanged(t *testing.T) {
	w := setUp(t)

	route, _ := url.Parse("http://localhost/api/")
	handler := newDummyHandler(t, "h1")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	h1 := w.routerHash
	r1 := w.router

	err = w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	h2 := w.routerHash
	r2 := w.router

	assert.Equal(t, h1, h2, "router hash should not change for same routes")
	assert.Equal(t, r1, r2, "router should not change for same routes")
}

func TestUpsertRouteShouldUpdateRouterIfRouteChanged(t *testing.T) {
	w := setUp(t)

	route, _ := url.Parse("http://localhost/api/")
	handler := newDummyHandler(t, "h1")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	h1 := w.routerHash
	r1 := w.router

	route, _ = url.Parse("http://localhost/api2/")

	err = w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	h2 := w.routerHash
	r2 := w.router

	assert.NotEqual(t, h1, h2, "router hash should change for different route URLs")
	assert.NotEqual(t, r1, r2, "router should change for different route URLs")
}

func TestUpsertRouteShouldUpdateRouterIfHandlerChanged(t *testing.T) {
	w := setUp(t)

	route, _ := url.Parse("http://localhost/api/")
	handler := newDummyHandler(t, "h1")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	h1 := w.routerHash
	r1 := w.router

	handler = newDummyHandler(t, "h2")

	err = w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	h2 := w.routerHash
	r2 := w.router

	assert.NotEqual(t, h1, h2, "router hash should change for different route handlers")
	assert.NotEqual(t, r1, r2, "router should change for different route handlers")
}

func TestWebServerShouldReturnNotFound(t *testing.T) {
	w := setUp(t)
	handler := newDummyHandler(t, "")

	route, _ := url.Parse("http://localhost/api/")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api2/test", nil)

	w.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
}

func _TestWebServerShouldFindWithoutSlash(t *testing.T) {
	w := setUp(t)
	handler := newDummyHandler(t, "")

	route, _ := url.Parse("http://localhost/api/")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api", nil)

	w.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
}

func TestWebServerShouldFindWithoutPath(t *testing.T) {
	s := setUp(t)
	handler := newDummyHandler(t, "")

	route, _ := url.Parse("http://localhost")

	err := s.UpsertRoute(route, handler)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/", nil)

	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestDeleteRouteShouldWorkForUnknownRoute(t *testing.T) {
	w := setUp(t)

	route, _ := url.Parse("http://localhost/api/")

	w.DeleteRoute(route)
}

func TestDeleteRouteShouldRemoveExistingRoute(t *testing.T) {
	w := setUp(t)

	route, _ := url.Parse("http://localhost/api")
	handler := newDummyHandler(t, "")

	err := w.UpsertRoute(route, handler)
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api", nil)

	w.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Result().StatusCode)

	route, _ = url.Parse("http://localhost/api")

	w.DeleteRoute(route)

	rw = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "http://localhost/api", nil)

	w.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
}

func TestNewWebServerOnBoundAddressShouldFail(t *testing.T) {
	_, err := NewWebServer(logging.NewLogrusLogger(logrus.New()), ":1234")
	assert.Nil(t, err)

	_, err = NewWebServer(logging.NewLogrusLogger(logrus.New()), ":1234")
	assert.NotNil(t, err)
}
