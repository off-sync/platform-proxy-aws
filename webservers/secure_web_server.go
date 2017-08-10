package webservers

import (
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/off-sync/platform-proxy-app/interfaces"
	"github.com/off-sync/platform-proxy-domain/frontends"
)

// SecureWebServer implements the Secure Web Server interface, extending the
// WebServer with certificate management.
type SecureWebServer struct {
	*WebServer

	certificatesLock sync.Mutex
	certificates     map[string]*tls.Certificate
}

// NewSecureWebServer creates a new secure web server listening on the provided
// address.
func NewSecureWebServer(log interfaces.Logger, addr string) (*SecureWebServer, error) {
	webServer, err := NewWebServer(log, addr)
	if err != nil {
		return nil, err
	}

	return &SecureWebServer{
		WebServer:    webServer,
		certificates: make(map[string]*tls.Certificate),
	}, nil
}

func (s *SecureWebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.WebServer.ServeHTTP(w, r)
}

// UpsertCertificate sets the certificate for the provided domain name.
func (s *SecureWebServer) UpsertCertificate(domainName string, cert *frontends.Certificate) error {
	s.certificatesLock.Lock()
	defer s.certificatesLock.Unlock()

	s.certificates[domainName] = cert.Certificate

	return nil
}
