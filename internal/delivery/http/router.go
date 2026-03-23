package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /reservations", h.Reserve)
	mux.HandleFunc("POST /reservations/{id}/return", h.Return)
	mux.HandleFunc("POST /reservations/{id}/deliver", h.Deliver)
	return mux
}
