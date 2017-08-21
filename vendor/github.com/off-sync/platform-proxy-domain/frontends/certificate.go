package frontends

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// Errors.
var (
	ErrInvalidCertificatePEM = errors.New("invalid certificate PEM")
	ErrInvalidKeyPEM         = errors.New("invalid key PEM")
)

// A Certificate wraps a certificate for use in a frontend.
type Certificate struct {
	Certificate *tls.Certificate
}

// NewCertificate creates a new Certificate using the provided PEM encoded
// certificate and key data.
func NewCertificate(certPEM, keyPEM []byte) (*Certificate, error) {
	var certs [][]byte

	var p *pem.Block
	for rest := certPEM; len(rest) > 0; {
		if p, rest = pem.Decode(rest); p != nil {
			certs = append(certs, p.Bytes)
		}

		if len(rest) == len(certPEM) {
			return nil, ErrInvalidCertificatePEM
		}
	}

	if len(certs) < 1 {
		return nil, fmt.Errorf("unable to decode certificate(s)")
	}

	leaf, err := x509.ParseCertificate(certs[len(certs)-1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse leaf certificate")
	}

	p, _ = pem.Decode(keyPEM)
	if p == nil {
		return nil, ErrInvalidKeyPEM
	}

	key, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %s", err)
	}

	return &Certificate{
		Certificate: &tls.Certificate{
			Certificate: certs,
			PrivateKey:  key,
			Leaf:        leaf,
		},
	}, nil
}

// ExpiresAt returns the time at which this certificate will expire.
func (cert *Certificate) ExpiresAt() time.Time {
	return cert.Certificate.Leaf.NotAfter.UTC()
}
