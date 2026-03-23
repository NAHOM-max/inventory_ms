package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors — use errors.Is to match these through wrapping.
var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrInsufficientReserve = errors.New("insufficient reserved amount")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrInvalidInput        = errors.New("invalid input")
)

// StockError carries the product ID alongside the stock sentinel.
type StockError struct {
	ProductID string
	Requested int
	Available int
	cause     error // ErrInsufficientStock or ErrInsufficientReserve
}

func NewStockError(productID string, requested, available int, cause error) *StockError {
	return &StockError{ProductID: productID, Requested: requested, Available: available, cause: cause}
}

func (e *StockError) Error() string {
	return fmt.Sprintf("product %s: requested %d, available %d: %s", e.ProductID, e.Requested, e.Available, e.cause)
}

func (e *StockError) Unwrap() error { return e.cause }

// TransitionError carries the order ID and the illegal from→to transition.
type TransitionError struct {
	OrderID string
	From    ReservationStatus
	To      ReservationStatus
}

func NewTransitionError(orderID string, from, to ReservationStatus) *TransitionError {
	return &TransitionError{OrderID: orderID, From: from, To: to}
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("order %s: cannot transition from %s to %s", e.OrderID, e.From, e.To)
}

func (e *TransitionError) Unwrap() error { return ErrInvalidTransition }

// ValidationError carries the field name and reason.
type ValidationError struct {
	Field  string
	Reason string
}

func NewValidationError(field, reason string) *ValidationError {
	return &ValidationError{Field: field, Reason: reason}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s — %s", e.Field, e.Reason)
}

func (e *ValidationError) Unwrap() error { return ErrInvalidInput }
