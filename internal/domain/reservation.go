package domain

import "time"

type ReservationStatus string

const (
	ReservationActive    ReservationStatus = "active"
	ReservationReturned  ReservationStatus = "returned"
	ReservationDelivered ReservationStatus = "delivered"
)

type Reservation struct {
	ID          string
	ItemID      string
	Quantity    int
	Status      ReservationStatus
	ReservedAt  time.Time
	ResolvedAt  *time.Time
}
