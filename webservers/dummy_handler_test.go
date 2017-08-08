package webservers

import (
	"fmt"
	"net/http"
	"testing"
)

type dummyHandler struct {
	t      *testing.T
	prefix string
}

func newDummyHandler(t *testing.T, prefix string) *dummyHandler {
	return &dummyHandler{
		t:      t,
		prefix: prefix,
	}
}

func (h *dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.t.Logf("[%s] serving HTTP: %v", h.prefix, r)

	fmt.Fprintf(w, "[%s] %s %s", h.prefix, r.Method, r.URL)
}
