package handler

import (
	"encoding/json"
	"net/http"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/dto"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/respond"
)

type EmployeeHandler struct {
	svc  domain.EmployeeService
	resp *respond.Responder
}

func NewEmployeeHandler(svc domain.EmployeeService, resp *respond.Responder) *EmployeeHandler {
	return &EmployeeHandler{svc: svc, resp: resp}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	deptID, err := pathID(r, "id")
	if err != nil {
		h.resp.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department id"})
		return
	}

	var req dto.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.resp.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	emp, err := h.svc.Create(r.Context(), deptID, req.FullName, req.Position, req.HiredAt)
	if err != nil {
		h.resp.Error(w, err)
		return
	}

	h.resp.JSON(w, http.StatusCreated, dto.EmployeeFromDomain(emp))
}
