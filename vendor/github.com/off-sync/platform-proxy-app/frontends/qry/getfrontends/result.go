package getfrontends

import "github.com/off-sync/platform-proxy-domain/frontends"

// Result specifies the output of the Query.
type Result struct {
	Frontends []*frontends.Frontend
}
