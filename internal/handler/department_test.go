package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/dto"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler/testhelper"
)

// ─── POST /departments ────────────────────────────────────────────────────────

func TestCreateDepartment_Success(t *testing.T) {
	s := testhelper.NewSetup()

	rec := s.Request(t, http.MethodPost, "/departments", map[string]any{
		"name": "Engineering",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.DepartmentResponse
	testhelper.DecodeResponse(t, rec, &resp)

	if resp.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if resp.Name != "Engineering" {
		t.Errorf("expected name %q, got %q", "Engineering", resp.Name)
	}
	if resp.ParentID != nil {
		t.Error("expected parent_id to be nil")
	}
}

func TestCreateDepartment_EmptyName(t *testing.T) {
	s := testhelper.NewSetup()

	rec := s.Request(t, http.MethodPost, "/departments", map[string]any{
		"name": "",
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateDepartment_WithParent(t *testing.T) {
	s := testhelper.NewSetup()

	// Создаём родительский отдел напрямую в моке
	parent := &domain.Department{Name: "Engineering"}
	_ = s.DeptRepo.Create(t.Context(), parent)

	rec := s.Request(t, http.MethodPost, "/departments", map[string]any{
		"name":      "Backend",
		"parent_id": parent.ID,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.DepartmentResponse
	testhelper.DecodeResponse(t, rec, &resp)

	if resp.ParentID == nil || *resp.ParentID != parent.ID {
		t.Errorf("expected parent_id %d, got %v", parent.ID, resp.ParentID)
	}
}

func TestCreateDepartment_ParentNotFound(t *testing.T) {
	s := testhelper.NewSetup()

	rec := s.Request(t, http.MethodPost, "/departments", map[string]any{
		"name":      "Backend",
		"parent_id": 999,
	})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ─── GET /departments/{id} ────────────────────────────────────────────────────

func TestGetDepartment_Success(t *testing.T) {
	s := testhelper.NewSetup()

	dept := &domain.Department{Name: "Engineering"}
	_ = s.DeptRepo.Create(t.Context(), dept)

	rec := s.Request(t, http.MethodGet, "/departments/1", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.DepartmentNodeResponse
	testhelper.DecodeResponse(t, rec, &resp)

	if resp.ID != dept.ID {
		t.Errorf("expected ID %d, got %d", dept.ID, resp.ID)
	}
}

func TestGetDepartment_NotFound(t *testing.T) {
	s := testhelper.NewSetup()

	rec := s.Request(t, http.MethodGet, "/departments/999", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ─── PATCH /departments/{id} ──────────────────────────────────────────────────

func TestUpdateDepartment_Rename(t *testing.T) {
	s := testhelper.NewSetup()

	dept := &domain.Department{Name: "Engeneering"} // опечатка специально
	_ = s.DeptRepo.Create(t.Context(), dept)

	newName := "Engineering"
	rec := s.Request(t, http.MethodPatch, "/departments/1", map[string]any{
		"name": newName,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.DepartmentResponse
	testhelper.DecodeResponse(t, rec, &resp)

	if resp.Name != newName {
		t.Errorf("expected name %q, got %q", newName, resp.Name)
	}
}

func TestUpdateDepartment_CycleDetection(t *testing.T) {
	s := testhelper.NewSetup()

	parent := &domain.Department{Name: "Parent"}
	_ = s.DeptRepo.Create(t.Context(), parent)

	child := &domain.Department{Name: "Child", ParentID: &parent.ID}
	_ = s.DeptRepo.Create(t.Context(), child)

	// Переопределяем ExistsInSubtree — имитируем что child входит в поддерево parent
	s.DeptRepo.ExistsInSubtreeFn = func(ctx context.Context, rootID, targetID int) (bool, error) {
		return rootID == parent.ID && targetID == child.ID, nil
	}

	// Пытаемся сделать parent дочерним для child — это цикл
	rec := s.Request(t, http.MethodPatch, "/departments/1", map[string]any{
		"parent_id": child.ID,
	})

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ─── DELETE /departments/{id} ─────────────────────────────────────────────────

func TestDeleteDepartment_Cascade(t *testing.T) {
	s := testhelper.NewSetup()

	dept := &domain.Department{Name: "ToDelete"}
	_ = s.DeptRepo.Create(t.Context(), dept)

	rec := s.Request(t, http.MethodDelete, "/departments/1?mode=cascade", nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if _, exists := s.DeptRepo.Departments[dept.ID]; exists {
		t.Error("expected department to be deleted")
	}
}

func TestDeleteDepartment_NoMode(t *testing.T) {
	s := testhelper.NewSetup()

	rec := s.Request(t, http.MethodDelete, "/departments/1", nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteDepartment_Reassign(t *testing.T) {
	s := testhelper.NewSetup()

	from := &domain.Department{Name: "From"}
	_ = s.DeptRepo.Create(t.Context(), from)

	to := &domain.Department{Name: "To"}
	_ = s.DeptRepo.Create(t.Context(), to)

	rec := s.Request(t, http.MethodDelete,
		"/departments/1?mode=reassign&reassign_to_department_id=2", nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}
