package dto

import (
	"time"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

// ─── Requests ────────────────────────────────────────────────────────────────

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name"`
	ParentID    *int    `json:"parent_id"`
	ClearParent bool    `json:"clear_parent"`
}

// ─── Responses ───────────────────────────────────────────────────────────────

type DepartmentResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ParentID  *int      `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type DepartmentNodeResponse struct {
	DepartmentResponse
	Employees []EmployeeResponse       `json:"employees,omitempty"`
	Children  []DepartmentNodeResponse `json:"children,omitempty"`
}

// ─── Converters ──────────────────────────────────────────────────────────────

func DepartmentFromDomain(d *domain.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:        d.ID,
		Name:      d.Name,
		ParentID:  d.ParentID,
		CreatedAt: d.CreatedAt,
	}
}

func DepartmentNodeFromDomain(node *domain.DepartmentNode) DepartmentNodeResponse {
	if node == nil || node.Department == nil {
		return DepartmentNodeResponse{}
	}

	resp := DepartmentNodeResponse{
		DepartmentResponse: DepartmentFromDomain(node.Department),
	}

	if len(node.Employees) > 0 {
		resp.Employees = make([]EmployeeResponse, len(node.Employees))
		for i, e := range node.Employees {
			resp.Employees[i] = EmployeeFromDomain(e)
		}
	}

	if len(node.Children) > 0 {
		resp.Children = make([]DepartmentNodeResponse, len(node.Children))
		for i, c := range node.Children {
			resp.Children[i] = DepartmentNodeFromDomain(c)
		}
	}

	return resp
}
