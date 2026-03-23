package domain

import "errors"

var (
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrInsufficientReserve = errors.New("insufficient reserved amount")
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidTransition   = errors.New("invalid status transition")
)
