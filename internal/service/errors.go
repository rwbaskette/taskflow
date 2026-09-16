package service

import "errors"

// Service-specific errors
var (
	ErrInvalidID          = errors.New("task ID is invalid or missing")
	ErrMissingBlockReason = errors.New("reason for blocking is required")
	ErrInvalidTimeout     = errors.New("timeout minutes must be a positive integer")
)
