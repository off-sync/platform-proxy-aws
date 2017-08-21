package startproxy

import (
	"context"
	"sync"
	"time"

	"github.com/off-sync/platform-proxy-app/interfaces"
)

// Model provides the input for the Start Proxy Command Execute method.
type Model struct {
	// Ctx is used to provide a means of stopping the created proxy once the
	// command is executed. This is achieved by closing the Done channel.
	Ctx context.Context

	// WaitGroup allows the Start Proxy command to signal to the calling process
	// that is finished cleaning up.
	WaitGroup *sync.WaitGroup

	// WebServer specifies the web server used for redirecting requests
	// for frontends with a certificate to the equivalent HTTPS URL. Frontends
	// without a certificate will be served on this web server directly.
	WebServer interfaces.WebServer

	// SecureWebServer specifies the web server used for frontends with a
	// certificate.
	SecureWebServer interfaces.SecureWebServer

	// LoadBalancer specifies the load balancer to use for making services
	// requests.
	LoadBalancer interfaces.LoadBalancer

	// PollingDuration defines the frequency at which the complete configuration
	// of the proxy is refreshed. This can be used when watchers are not
	// available, or when watchers are not reliable (i.e. change events could be
	// missed). Polling is disabled when this duration is set to the zero value.
	PollingDuration time.Duration
}
