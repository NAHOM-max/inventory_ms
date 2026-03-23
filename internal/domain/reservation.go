package domain

type ReservationStatus string

const (
	StatusReserved  ReservationStatus = "RESERVED"
	StatusReturned  ReservationStatus = "RETURNED"
	StatusDelivered ReservationStatus = "DELIVERED"
)

type ReservationItem struct {
	ProductID  string
	Amount     int
	TotalPrice float64
}

type Reservation struct {
	OrderID    string
	Status     ReservationStatus
	TotalPrice float64
	Items      []ReservationItem
}
