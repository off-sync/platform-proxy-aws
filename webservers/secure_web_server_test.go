package webservers

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/off-sync/platform-proxy-app/infra/logging"
	"github.com/stretchr/testify/assert"
)

func setUpSecureWebServer(t *testing.T) *SecureWebServer {
	s, err := NewSecureWebServer(logging.NewLogrusLogger(logrus.New()), ":0")
	assert.Nil(t, err)

	return s
}

func TestNewSecureWebServer(t *testing.T) {
	s := setUpSecureWebServer(t)
	assert.NotNil(t, s)
}

func TestSecureWebServerShouldServeHTTP(t *testing.T) {
	s := setUpSecureWebServer(t)

	route, _ := url.Parse("http://localhost/api")
	handler := newDummyHandler(t, "")

	s.UpsertRoute(route, handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/test", nil)

	s.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)

	assert.Equal(t, "[] GET http://localhost/api/test", string(body))
}

func TestSecureWebServerUpsertCertificate(t *testing.T) {
	s := setUpSecureWebServer(t)

	cert := getTestCert()
	assert.NotNil(t, cert)

	err := s.UpsertCertificate("localhost", cert)
	assert.Nil(t, err)
}
