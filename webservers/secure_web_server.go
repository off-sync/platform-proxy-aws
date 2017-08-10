package webservers

import (
	"crypto/tls"
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
	c := newCertificateManager()

	s := &SecureWebServer{
		routeManager:       newRouteManager(log),
		certificateManager: c,
		server: &http.Server{
			Addr: addr,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
				GetCertificate: c.getCertificate,
			},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		},
	}

	s.server.Handler = s

	// start the https server
	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil {
			log.WithError(err).Error("listening and serving TLS")
		}
	}()

	return s
}
