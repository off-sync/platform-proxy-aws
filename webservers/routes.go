package webservers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
)

type routes map[string]*route

func (r routes) hash() string {
	buf := &bytes.Buffer{}

	var keys []string
	for key := range r {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		route := r[key]
		fmt.Fprintf(buf, "%s:%p", key, route.handler)
	}

	h := sha256.Sum256(buf.Bytes())
	return string(h[:])
}
