package dto

import (
	"time"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

// ─── Requests ────────────────────────────────────────────────────────────────

type CreateEmployeeRequest struct {
	FullName string     `json:"full_name"`
	Position string     `json:"position"`
	HiredAt  *time.Time `json:"hired_at"`
}

// ─── Responses ───────────────────────────────────────────────────────────────

type EmployeeResponse struct {
	ID           int        `json:"id"`
	DepartmentID int        `json:"department_id"`
	FullName     string     `json:"full_name"`
	Position     string     `json:"position"`
	HiredAt      *time.Time `json:"hired_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ─── Converters ──────────────────────────────────────────────────────────────

func EmployeeFromDomain(e *domain.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:           e.ID,
		DepartmentID: e.DepartmentID,
		FullName:     e.FullName,
		Position:     e.Position,
		HiredAt:      e.HiredAt,
		CreatedAt:    e.CreatedAt,
	}
}
