package respond

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

type Responder struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Responder {
	return &Responder{log: log}
}

func (r *Responder) JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		r.log.Error("failed to encode response", slog.String("err", err.Error()))
	}
}

func (r *Responder) Error(w http.ResponseWriter, err error) {
	code, msg := httpError(err)
	r.JSON(w, code, map[string]string{"error": msg})
}

func httpError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
