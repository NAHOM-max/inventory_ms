package http

import (
	"encoding/json"
	"net/http"

	"inventory_ms/internal/usecase"
)

type Handler struct {
	reserve *usecase.ReserveUseCase
	ret     *usecase.ReturnUseCase
	deliver *usecase.DeliverUseCase
}

func NewHandler(r *usecase.ReserveUseCase, ret *usecase.ReturnUseCase, d *usecase.DeliverUseCase) *Handler {
	return &Handler{reserve: r, ret: ret, deliver: d}
}

func (h *Handler) Reserve(w http.ResponseWriter, r *http.Request) {
	// TODO: decode request, call h.reserve.Execute
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) Return(w http.ResponseWriter, r *http.Request) {
	// TODO: decode request, call h.ret.Execute
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) Deliver(w http.ResponseWriter, r *http.Request) {
	// TODO: decode request, call h.deliver.Execute
	w.WriteHeader(http.StatusAccepted)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
