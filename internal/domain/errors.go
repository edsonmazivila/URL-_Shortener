package domain

import "errors"

var (
	// ErrURLNotFound is returned when a URL cannot be found.
	ErrURLNotFound = errors.New("url not found")

	// ErrURLExpired is returned when a URL has expired.
	ErrURLExpired = errors.New("url has expired")

	// ErrInvalidURL is returned when the provided URL is invalid.
	ErrInvalidURL = errors.New("invalid url")

	// ErrShortCodeAlreadyExists is returned when a short code is already in use.
	ErrShortCodeAlreadyExists = errors.New("short code already exists")

	// ErrInvalidShortCode is returned when the short code format is invalid.
	ErrInvalidShortCode = errors.New("invalid short code")
)
