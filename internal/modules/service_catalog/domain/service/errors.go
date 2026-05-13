package service

import "errors"

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrDuplicateServiceName = errors.New("service name already exists")
	ErrInvalidServiceName   = errors.New("service name must be between 3 and 100 characters")
	ErrInvalidDescription   = errors.New("description must be between 10 and 500 characters")
	ErrInvalidPrice         = errors.New("price must be greater than zero")
	ErrInvalidServiceID     = errors.New("invalid service ID format")
)
