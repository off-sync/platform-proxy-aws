package webservers

import (
	"crypto/tls"
	"sync"

	"github.com/off-sync/platform-proxy-domain/frontends"
)

type certificateManager struct {
	certificatesLock sync.Mutex
	certificates     map[string]*tls.Certificate
}

func newCertificateManager() *certificateManager {
	return &certificateManager{
		certificates: make(map[string]*tls.Certificate),
	}
}

// UpsertCertificate sets the certificate for the provided domain name.
func (m *certificateManager) UpsertCertificate(domainName string, cert *frontends.Certificate) error {
	m.certificatesLock.Lock()
	defer m.certificatesLock.Unlock()

	m.certificates[domainName] = cert.Certificate

	return nil
}
