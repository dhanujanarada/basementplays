package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

type publicAPI struct {
	logger       *slog.Logger
	availability *service.Availability
	holds        *service.Holds
	orders       *service.Orders
	tickets      *service.Tickets
	writeLimiter *rateLimiter
}

func (a *publicAPI) mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/public/shows/{id}", a.showSummary)
	mux.HandleFunc("GET /v1/public/shows/{id}/seats", a.showSeats)
	mux.HandleFunc("POST /v1/public/shows/{id}/holds", a.createHold)
	mux.HandleFunc("DELETE /v1/public/orders/{id}", a.releaseOrder)
	mux.HandleFunc("POST /v1/public/orders/{id}/checkout", a.checkoutOrder)
	mux.HandleFunc("GET /v1/public/orders/{id}", a.getOrder)
	mux.HandleFunc("GET /v1/public/tickets/{token}", a.getTicket)
}

func (a *publicAPI) showSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := a.availability.Summary(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (a *publicAPI) showSeats(w http.ResponseWriter, r *http.Request) {
	seatMap, err := a.availability.SeatMap(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	a.writeETagJSON(w, r, seatMap)
}

func (a *publicAPI) createHold(w http.ResponseWriter, r *http.Request) {
	if !a.allowWrite(w, r) {
		return
	}

	var input struct {
		Seats []string `json:"seats"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	hold, replayed, err := a.holds.Create(
		r.Context(),
		r.PathValue("id"),
		input.Seats,
		r.Header.Get("Idempotency-Key"),
	)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
		w.Header().Set("Idempotent-Replayed", "true")
	}

	writeJSON(w, status, hold)
}

func (a *publicAPI) releaseOrder(w http.ResponseWriter, r *http.Request) {
	if !a.allowWrite(w, r) {
		return
	}

	token, ok := orderToken(w, r)
	if !ok {
		return
	}

	if err := a.holds.Release(r.Context(), r.PathValue("id"), token); err != nil {
		a.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *publicAPI) checkoutOrder(w http.ResponseWriter, r *http.Request) {
	if !a.allowWrite(w, r) {
		return
	}

	token, ok := orderToken(w, r)
	if !ok {
		return
	}

	var input service.CheckoutInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	order, err := a.orders.Checkout(
		r.Context(),
		r.PathValue("id"),
		token,
		input,
		r.Header.Get("Idempotency-Key"),
	)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (a *publicAPI) getOrder(w http.ResponseWriter, r *http.Request) {
	token, ok := orderToken(w, r)
	if !ok {
		return
	}

	order, err := a.orders.Get(r.Context(), r.PathValue("id"), token)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (a *publicAPI) getTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := a.tickets.Lookup(r.Context(), r.PathValue("token"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (a *publicAPI) allowWrite(w http.ResponseWriter, r *http.Request) bool {
	if a.writeLimiter == nil {
		return true
	}

	ok, retry := a.writeLimiter.allow(clientIP(r))
	if !ok {
		writeRateLimited(w, retry)
		return false
	}

	return true
}

func orderToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	token := strings.TrimSpace(r.Header.Get("X-Order-Token"))
	if token == "" {
		writeAPIError(w, http.StatusUnauthorized, "missing_order_token", "X-Order-Token is required")
		return "", false
	}

	return token, true
}

func (a *publicAPI) writeETagJSON(w http.ResponseWriter, r *http.Request, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")

	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func etagMatches(header, etag string) bool {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, "W/")
		part = strings.TrimSpace(part)
		if part == "*" || part == etag {
			return true
		}
	}

	return false
}

func (a *publicAPI) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		writeAPIError(w, http.StatusBadRequest, "validation_error", validation.Error())

	case errors.Is(err, store.ErrShowNotFound), errors.Is(err, store.ErrOrderNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", "not found")

	case errors.Is(err, store.ErrSalesClosed):
		writeAPIError(w, http.StatusConflict, "sales_closed", "tickets are not on sale for this show")

	case errors.Is(err, store.ErrSeatTaken):
		writeAPIError(w, http.StatusConflict, "seat_taken", "one or more selected seats are no longer available")

	case errors.Is(err, store.ErrUnknownSeat):
		writeAPIError(w, http.StatusBadRequest, "unknown_seat", "one or more seats do not exist for this show")

	case errors.Is(err, store.ErrIdempotencyMismatch):
		writeAPIError(w, http.StatusConflict, "idempotency_mismatch", "Idempotency-Key was already used with different seats")

	case errors.Is(err, store.ErrOrderTokenMismatch):
		writeAPIError(w, http.StatusForbidden, "invalid_order_token", "X-Order-Token is invalid")

	case errors.Is(err, store.ErrOrderNotPending):
		writeAPIError(w, http.StatusConflict, "order_not_pending", "only pending holds can be released")

	case errors.Is(err, store.ErrHoldExpired):
		writeAPIError(w, http.StatusConflict, "hold_expired", "the hold has expired; choose seats again")

	case errors.Is(err, store.ErrOrderNotCheckoutable):
		writeAPIError(w, http.StatusConflict, "order_not_checkoutable", "the order is no longer awaiting checkout")

	case errors.Is(err, store.ErrOrderNeedsReview):
		writeAPIError(w, http.StatusConflict, "order_needs_review", "the order is awaiting manual review")

	case errors.Is(err, service.ErrInvalidTicketToken), errors.Is(err, store.ErrTicketNotFound):
		writeAPIError(w, http.StatusNotFound, "ticket_not_found", "ticket not found")

	default:
		a.logger.Error("public request failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestIDFrom(r.Context()),
		)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
