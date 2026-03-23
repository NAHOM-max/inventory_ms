package domain

import "time"

type InventoryStatus string

const (
	StatusAvailable InventoryStatus = "available"
	StatusReserved  InventoryStatus = "reserved"
	StatusDelivered InventoryStatus = "delivered"
)

type InventoryItem struct {
	ID        string
	Name      string
	Quantity  int
	Status    InventoryStatus
	UpdatedAt time.Time
}
