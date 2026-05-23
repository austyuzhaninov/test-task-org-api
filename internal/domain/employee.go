package domain

import (
	"context"
	"time"
)

// Employee — чистая бизнес-сущность.
type Employee struct {
	ID           int
	DepartmentID int
	FullName     string
	Position     string
	HiredAt      *time.Time
	CreatedAt    time.Time
}

// EmployeeRepository — интерфейс принадлежит домену.
type EmployeeRepository interface {
	Create(ctx context.Context, e *Employee) error
	ListByDepartment(ctx context.Context, departmentID int) ([]*Employee, error)
}

// EmployeeService — интерфейс бизнес-логики, используется хендлером.
type EmployeeService interface {
	Create(ctx context.Context, departmentID int, fullName, position string, hiredAt *time.Time) (*Employee, error)
}
