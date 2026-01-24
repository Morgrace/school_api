package models

import "errors"

var (
	// 404: Resource does not exist
	ErrNotFound = errors.New("resource not found")

	// 409: Resource already exists (e.g. duplicate email)
	ErrConflict = errors.New("resource already exists")

	// 400: Input is bad (e.g. valid JSON but invalid business logic)
	ErrInvalidInput = errors.New("invalid input data")

	// 401: Authentication failed
	ErrUnauthorized = errors.New("unauthorized")

	// 500: Explicit system failure (optional, usually implied by unknown errors)
	ErrInternal = errors.New("internal system error")
)
