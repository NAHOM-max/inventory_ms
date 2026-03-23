package domain


type Inventory struct {
	ProductID       string
	ProductName     string
	ProductPrice    float64
	ProductWeight   float64
	AvailableAmount int
	ReservedAmount  int
}

// CanReserve reports whether the requested amount can be reserved.
func (inv *Inventory) CanReserve(amount int) bool {
	return inv.AvailableAmount-inv.ReservedAmount >= amount
}

// Reserve increases the reserved amount.
func (inv *Inventory) Reserve(amount int) error {
	if !inv.CanReserve(amount) {
		return ErrInsufficientStock
	}
	inv.ReservedAmount += amount
	return nil
}

// Release decreases the reserved amount (e.g. on return).
func (inv *Inventory) Release(amount int) error {
	if inv.ReservedAmount < amount {
		return ErrInsufficientReserve
	}
	inv.ReservedAmount -= amount
	return nil
}

// Deliver decreases both available and reserved amounts (stock leaves warehouse).
func (inv *Inventory) Deliver(amount int) error {
	if inv.ReservedAmount < amount {
		return ErrInsufficientReserve
	}
	if inv.AvailableAmount < amount {
		return ErrInsufficientStock
	}
	inv.ReservedAmount -= amount
	inv.AvailableAmount -= amount
	return nil
}
