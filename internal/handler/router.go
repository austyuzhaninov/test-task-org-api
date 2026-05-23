package handler

import (
	"fmt"
	"net/http"

	"github.com/austyuzhaninov/test-task-org-api/internal/middleware"
)

func NewRouter(dept *DepartmentHandler, emp *EmployeeHandler, log *middleware.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Departments
	mux.HandleFunc("POST /departments", dept.Create)
	mux.HandleFunc("GET /departments/{id}", dept.GetByID)
	mux.HandleFunc("PATCH /departments/{id}", dept.Update)
	mux.HandleFunc("DELETE /departments/{id}", dept.Delete)

	// Employees
	mux.HandleFunc("POST /departments/{id}/employees/", emp.Create)

	return log.Middleware(mux)
}
