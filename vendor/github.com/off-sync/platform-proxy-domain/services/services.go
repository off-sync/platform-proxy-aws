// Package services contains the domain models for the Service aggregate root.
package services

import "errors"

// Errors
var (
	ErrServiceNameMissing = errors.New("service name missing")
	ErrServerURLsMissing  = errors.New("server URLs missing")
)
