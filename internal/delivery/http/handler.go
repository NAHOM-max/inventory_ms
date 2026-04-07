package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"inventory_ms/internal/domain"
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
	var req reserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	items := make([]usecase.ReserveItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = usecase.ReserveItemInput{ProductID: it.ProductID, Amount: it.Amount}
	}

	err := h.reserve.Execute(r.Context(), usecase.ReserveInput{
		OrderID: req.OrderID,
		Items:   items,
	})
	if err != nil {
		writeError(w, domainStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, reserveResponse{ReservationID: "res-" + req.OrderID})
}

func (h *Handler) Return(w http.ResponseWriter, r *http.Request) {
	var req orderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ret.Execute(r.Context(), req.OrderID); err != nil {
		writeError(w, domainStatus(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Deliver(w http.ResponseWriter, r *http.Request) {
	var req orderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.deliver.Execute(r.Context(), req.OrderID); err != nil {
		writeError(w, domainStatus(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// domainStatus maps domain sentinel errors to HTTP status codes.
func domainStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInsufficientStock),
		errors.Is(err, domain.ErrInsufficientReserve),
		errors.Is(err, domain.ErrInvalidTransition):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
