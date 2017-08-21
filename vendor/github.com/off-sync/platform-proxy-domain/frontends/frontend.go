package frontends

import "net/url"

// A Frontend defines how a Service is exposed through the Platform Proxy.
type Frontend struct {
	Name        string
	URL         *url.URL
	Certificate *Certificate
	ServiceName string
}

// NewFrontend returns a Frontend by parsing the provided url. An error is
// returned if the url cannot be parsed.
func NewFrontend(name, frontendURL string, cert *Certificate, serviceName string) (*Frontend, error) {
	if name == "" {
		return nil, ErrFrontendNameMissing
	}

	if frontendURL == "" {
		return nil, ErrFrontendURLMissing
	}

	u, err := url.Parse(frontendURL)
	if err != nil {
		return nil, err
	}

	if serviceName == "" {
		return nil, ErrServiceNameMissing
	}

	return &Frontend{
		Name:        name,
		URL:         u,
		Certificate: cert,
		ServiceName: serviceName,
	}, nil
}
