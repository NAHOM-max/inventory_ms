package http

type reserveRequest struct {
	OrderID string             `json:"order_id"`
	Items   []reserveItemInput `json:"items"`
}

type reserveItemInput struct {
	ProductID string `json:"product_id"`
	Amount    int    `json:"amount"`
}

type orderRequest struct {
	OrderID string `json:"order_id"`
}

type reserveResponse struct {
	ReservationID string `json:"reservation_id"`
}

type returnResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}
