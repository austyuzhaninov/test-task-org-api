package handler

import (
	"fmt"
	"net/http"
	"strconv"
)

// pathID извлекает целочисленный {id} из пути через PathValue (Go 1.22+).
func pathID(r *http.Request, name string) (int, error) {
	raw := r.PathValue(name)
	if raw == "" {
		return 0, fmt.Errorf("missing path param %q", name)
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return 0, fmt.Errorf("invalid path param %q: %s", name, raw)
	}
	return v, nil
}

// queryInt читает query-параметр как int, возвращает defaultVal если отсутствует.
func queryInt(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return v
}

// queryBool читает query-параметр как bool, возвращает defaultVal если отсутствует.
func queryBool(r *http.Request, key string, defaultVal bool) bool {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultVal
	}
	return v
}
