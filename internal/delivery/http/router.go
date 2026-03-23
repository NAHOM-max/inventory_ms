package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /reserve", h.Reserve)
	mux.HandleFunc("POST /return", h.Return)
	mux.HandleFunc("POST /deliver", h.Deliver)
	return mux
}
