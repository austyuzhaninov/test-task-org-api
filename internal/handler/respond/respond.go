package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

// JSON сериализует v и пишет в ответ с нужным статусом.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error маппит доменные ошибки на HTTP-статусы и пишет JSON-ответ.
func Error(w http.ResponseWriter, err error) {
	code, msg := httpError(err)
	JSON(w, code, map[string]string{"error": msg})
}

// httpError определяет HTTP-статус по типу ошибки.
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
