package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const maxJSONBody = 1 << 20 // 1 MiB

type internalAPI struct {
	logger  *slog.Logger
	plans   *service.Plans
	shows   *service.Shows
	orders  *service.Orders
	scanner *service.Scanner
}

func (a *internalAPI) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /seat-plans", a.createPlan)
	mux.HandleFunc("GET /seat-plans", a.listPlans)
	mux.HandleFunc("GET /seat-plans/{key}", a.getPlan)
	mux.HandleFunc("PATCH /seat-plans/{key}/labels", a.updatePlanLabels)
	mux.HandleFunc("POST /seat-plans/{key}/duplicate", a.duplicatePlan)
	mux.HandleFunc("PUT /shows/{id}", a.upsertShow)
	if a.scanner != nil {
		mux.HandleFunc("POST /shows/{id}/scanner-token", a.issueScannerToken)
		mux.HandleFunc("DELETE /shows/{id}/scanner-token", a.revokeScannerToken)
	}
	mux.HandleFunc("POST /orders", a.createManualOrder)
	mux.HandleFunc("POST /orders/{id}/mark-paid", a.markOrderPaid)
	mux.HandleFunc("POST /orders/{id}/cancel", a.cancelOrder)
	mux.HandleFunc("GET /shows/{id}/orders", a.listShowOrders)

	return mux
}

func (a *internalAPI) createPlan(w http.ResponseWriter, r *http.Request) {
	var input service.CreatePlanInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	plan, err := a.plans.Create(r.Context(), input)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, plan)
}

func (a *internalAPI) listPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := a.plans.List(r.Context())
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"plans": plans})
}

func (a *internalAPI) getPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := a.plans.Get(r.Context(), r.PathValue("key"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

func (a *internalAPI) updatePlanLabels(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Labels map[string]*string `json:"labels"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if input.Labels == nil {
		input.Labels = map[string]*string{}
	}

	plan, err := a.plans.UpdateLabels(r.Context(), r.PathValue("key"), input.Labels)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

func (a *internalAPI) duplicatePlan(w http.ResponseWriter, r *http.Request) {
	var input service.DuplicatePlanInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	plan, err := a.plans.Duplicate(r.Context(), r.PathValue("key"), input)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, plan)
}

func (a *internalAPI) upsertShow(w http.ResponseWriter, r *http.Request) {
	var input service.UpsertShowInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	show, err := a.shows.Upsert(r.Context(), r.PathValue("id"), input)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, show)
}

func (a *internalAPI) issueScannerToken(w http.ResponseWriter, r *http.Request) {
	token, err := a.scanner.IssueToken(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, token)
}

func (a *internalAPI) revokeScannerToken(w http.ResponseWriter, r *http.Request) {
	if err := a.scanner.RevokeToken(r.Context(), r.PathValue("id")); err != nil {
		a.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *internalAPI) createManualOrder(w http.ResponseWriter, r *http.Request) {
	var input service.ManualOrderInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	order, err := a.orders.CreateManual(r.Context(), input)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (a *internalAPI) markOrderPaid(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PaymentMethod *string `json:"payment_method"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	order, err := a.orders.MarkPaid(r.Context(), r.PathValue("id"), input.PaymentMethod)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (a *internalAPI) cancelOrder(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	order, err := a.orders.Cancel(r.Context(), r.PathValue("id"), input.Reason)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (a *internalAPI) listShowOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := a.orders.List(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"orders": orders})
}

func (a *internalAPI) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		writeAPIError(w, http.StatusBadRequest, "validation_error", validation.Error())

	case errors.Is(err, store.ErrNotFound),
		errors.Is(err, store.ErrPlanNotFound),
		errors.Is(err, store.ErrShowNotFound),
		errors.Is(err, store.ErrOrderNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", "not found")

	case errors.Is(err, store.ErrPlanKeyExists):
		writeAPIError(w, http.StatusConflict, "plan_key_exists", "a seat plan with that key already exists")

	case errors.Is(err, store.ErrPlanInUse):
		writeAPIError(w, http.StatusConflict, "plan_in_use", "seat plans cannot be edited while a show uses them")

	case errors.Is(err, store.ErrCannotRelink):
		writeAPIError(w, http.StatusConflict, "show_in_use", "show has orders or unavailable seats; it cannot be relinked")

	case errors.Is(err, store.ErrSalesClosed):
		writeAPIError(w, http.StatusConflict, "sales_closed", "tickets are not on sale for this show")

	case errors.Is(err, store.ErrSeatTaken):
		writeAPIError(w, http.StatusConflict, "seat_taken", "one or more selected seats are no longer available")

	case errors.Is(err, store.ErrUnknownSeat):
		writeAPIError(w, http.StatusBadRequest, "unknown_seat", "one or more seats do not exist for this show")

	case errors.Is(err, store.ErrIdempotencyMismatch):
		writeAPIError(w, http.StatusConflict, "idempotency_mismatch", "Idempotency-Key was already used for this show")

	case errors.Is(err, store.ErrOrderNotPayable):
		writeAPIError(w, http.StatusConflict, "order_not_payable", "the order cannot be marked paid")

	case errors.Is(err, store.ErrOrderNotCancellable):
		writeAPIError(w, http.StatusConflict, "order_not_cancellable", "the order cannot be cancelled")

	case errors.Is(err, store.ErrTicketsInUse):
		writeAPIError(w, http.StatusConflict, "tickets_in_use", "the order has tickets that were already used")

	default:
		a.logger.Error("internal request failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestIDFrom(r.Context()),
		)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain exactly one JSON object")
	}

	return nil
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func withInternalAuth(token string, next http.Handler) http.Handler {
	expected := []byte(token)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "

		if token == "" || !strings.HasPrefix(header, prefix) {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token")
			return
		}

		got := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if subtle.ConstantTimeCompare([]byte(got), expected) != 1 {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
