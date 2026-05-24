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

// Create godoc
// @Summary     Добавить сотрудника в подразделение
// @Tags        employees
// @Accept      json
// @Produce     json
// @Param       id   path int                      true "ID подразделения"
// @Param       body body dto.CreateEmployeeRequest true "Данные сотрудника"
// @Success     201 {object} dto.EmployeeResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse "Подразделение не найдено"
// @Failure     422 {object} dto.ErrorResponse "Невалидные данные"
// @Router      /departments/{id}/employees [post]
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	deptID, err := pathID(r, "id")
	if err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid department id"})
		return
	}

	var req dto.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid json"})
		return
	}

	emp, err := h.svc.Create(r.Context(), deptID, req.FullName, req.Position, req.HiredAt)
	if err != nil {
		h.resp.Error(w, err)
		return
	}

	h.resp.JSON(w, http.StatusCreated, dto.EmployeeFromDomain(emp))
}
