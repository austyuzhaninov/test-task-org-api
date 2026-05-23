package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

type employeeModel struct {
	ID           int    `gorm:"primaryKey;autoIncrement"`
	DepartmentID int    `gorm:"not null"`
	FullName     string `gorm:"not null"`
	Position     string `gorm:"not null"`
	HiredAt      *time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (employeeModel) TableName() string { return "employees" }

// ────────────────────────────────────────────────────────────────────────────

type EmployeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{db: db}
}

func toEmpModel(e *domain.Employee) *employeeModel {
	return &employeeModel{
		ID:           e.ID,
		DepartmentID: e.DepartmentID,
		FullName:     e.FullName,
		Position:     e.Position,
		HiredAt:      e.HiredAt,
		CreatedAt:    e.CreatedAt,
	}
}

func toEmpDomain(m *employeeModel) *domain.Employee {
	return &domain.Employee{
		ID:           m.ID,
		DepartmentID: m.DepartmentID,
		FullName:     m.FullName,
		Position:     m.Position,
		HiredAt:      m.HiredAt,
		CreatedAt:    m.CreatedAt,
	}
}

// ────────────────────────────────────────────────────────────────────────────

func (r *EmployeeRepo) Create(ctx context.Context, e *domain.Employee) error {
	m := toEmpModel(e)

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapDBError(err)
	}

	e.ID = m.ID
	e.CreatedAt = m.CreatedAt
	return nil
}

func (r *EmployeeRepo) ListByDepartment(ctx context.Context, departmentID int) ([]*domain.Employee, error) {
	var models []employeeModel
	if err := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Order("created_at").
		Find(&models).Error; err != nil {
		return nil, mapDBError(err)
	}

	result := make([]*domain.Employee, len(models))
	for i := range models {
		result[i] = toEmpDomain(&models[i])
	}
	return result, nil
}
