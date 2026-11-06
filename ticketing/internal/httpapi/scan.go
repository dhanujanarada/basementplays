package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
)

// scannerTokenHeader carries the per-show credential issued by the internal
// scanner-token endpoint.
const scannerTokenHeader = "X-Scanner-Token"

type scanAPI struct {
	logger  *slog.Logger
	scanner *service.Scanner
}

func (a *scanAPI) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /admit", a.admit)
	return mux
}

func (a *scanAPI) admit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QRToken string `json:"qr_token"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	result, err := a.scanner.Admit(r.Context(), r.Header.Get(scannerTokenHeader), input.QRToken)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (a *scanAPI) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		validation *domain.ValidationError
		notIssued  *service.TicketNotIssuedError
	)

	switch {
	case errors.As(err, &validation):
		writeAPIError(w, http.StatusBadRequest, "validation_error", validation.Error())

	case errors.Is(err, service.ErrInvalidScannerToken):
		writeAPIError(w, http.StatusUnauthorized, "invalid_scanner_token", "missing or invalid scanner token")

	case errors.Is(err, service.ErrInvalidTicketToken):
		writeAPIError(w, http.StatusNotFound, "ticket_not_found", "ticket not found")

	case errors.Is(err, service.ErrTicketWrongShow):
		writeAPIError(w, http.StatusForbidden, "wrong_show", "the ticket belongs to a different show")

	case errors.As(err, &notIssued):
		code := "ticket_not_admittable"
		switch notIssued.Status {
		case "used":
			code = "ticket_already_used"
		case "void":
			code = "ticket_void"
		}
		writeAPIError(w, http.StatusConflict, code, "the ticket is "+notIssued.Status)

	default:
		a.logger.Error("scan request failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestIDFrom(r.Context()),
		)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
