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
	svc domain.DepartmentService
}

func NewDepartmentHandler(svc domain.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

// POST /departments/
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	dept, err := h.svc.Create(r.Context(), req.Name, req.ParentID)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, dto.DepartmentFromDomain(dept))
}

// GET /departments/{id}
func (h *DepartmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	depth := queryInt(r, "depth", 1)
	if depth > 5 {
		depth = 5
	}
	includeEmployees := queryBool(r, "include_employees", true)

	node, err := h.svc.GetByID(r.Context(), id, depth, includeEmployees)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, dto.DepartmentNodeFromDomain(node))
}

// PATCH /departments/{id}
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	dept, err := h.svc.Update(r.Context(), id, req.Name, req.ParentID, req.ClearParent)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, dto.DepartmentFromDomain(dept))
}

// DELETE /departments/{id}
func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "mode is required (cascade or reassign)"})
		return
	}

	var reassignTo *int
	if raw := r.URL.Query().Get("reassign_to_department_id"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			respond.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid reassign_to_department_id"})
			return
		}
		reassignTo = &v
	}

	if err := h.svc.Delete(r.Context(), id, mode, reassignTo); err != nil {
		respond.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
