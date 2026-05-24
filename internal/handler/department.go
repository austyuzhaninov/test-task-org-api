package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/dto"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/respond"
)

type DepartmentHandler struct {
	svc  domain.DepartmentService
	resp *respond.Responder
}

func NewDepartmentHandler(svc domain.DepartmentService, resp *respond.Responder) *DepartmentHandler {
	return &DepartmentHandler{svc: svc, resp: resp}
}

// Create godoc
// @Summary     Создать подразделение
// @Tags        departments
// @Accept      json
// @Produce     json
// @Param       body body dto.CreateDepartmentRequest true "Данные подразделения"
// @Success     201 {object} dto.DepartmentResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse "Родительское подразделение не найдено"
// @Failure     409 {object} dto.ErrorResponse "Подразделение с таким именем уже существует"
// @Failure     422 {object} dto.ErrorResponse "Невалидные данные"
// @Router      /departments [post]
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid json"})
		return
	}

	dept, err := h.svc.Create(r.Context(), req.Name, req.ParentID)
	if err != nil {
		h.resp.Error(w, err)
		return
	}

	h.resp.JSON(w, http.StatusCreated, dto.DepartmentFromDomain(dept))
}

// GetByID godoc
// @Summary     Получить подразделение
// @Description Возвращает подразделение с сотрудниками и деревом дочерних подразделений
// @Tags        departments
// @Produce     json
// @Param       id                path  int  true  "ID подразделения"
// @Param       depth             query int  false "Глубина дерева (1-5, по умолчанию 1)"
// @Param       include_employees query bool false "Включить сотрудников (по умолчанию true)"
// @Success     200 {object} dto.DepartmentNodeResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Router      /departments/{id} [get]
func (h *DepartmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	depth := queryInt(r, "depth", 1)
	if depth > 5 {
		depth = 5
	}
	includeEmployees := queryBool(r, "include_employees", true)

	node, err := h.svc.GetByID(r.Context(), id, depth, includeEmployees)
	if err != nil {
		h.resp.Error(w, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, dto.DepartmentNodeFromDomain(node))
}

// Update godoc
// @Summary     Обновить подразделение
// @Description Переименовать или переместить подразделение в другой родительский отдел
// @Tags        departments
// @Accept      json
// @Produce     json
// @Param       id   path int                         true "ID подразделения"
// @Param       body body dto.UpdateDepartmentRequest true "Данные для обновления"
// @Success     200 {object} dto.DepartmentResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Цикл в дереве или дубль имени"
// @Failure     422 {object} dto.ErrorResponse
// @Router      /departments/{id} [patch]
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid json"})
		return
	}

	dept, err := h.svc.Update(r.Context(), id, req.Name, req.ParentID, req.ClearParent)
	if err != nil {
		h.resp.Error(w, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, dto.DepartmentFromDomain(dept))
}

// Delete godoc
// @Summary     Удалить подразделение
// @Tags        departments
// @Produce     json
// @Param       id                        path  int    true  "ID подразделения"
// @Param       mode                      query string true  "Режим удаления: cascade или reassign"
// @Param       reassign_to_department_id query int    false "ID отдела для перевода сотрудников (обязателен при mode=reassign)"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Есть дочерние отделы при mode=reassign"
// @Router      /departments/{id} [delete]
func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "mode is required (cascade or reassign)"})
		return
	}

	var reassignTo *int
	if raw := r.URL.Query().Get("reassign_to_department_id"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			h.resp.JSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid reassign_to_department_id"})
			return
		}
		reassignTo = &v
	}

	if err := h.svc.Delete(r.Context(), id, mode, reassignTo); err != nil {
		h.resp.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
