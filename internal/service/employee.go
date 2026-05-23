package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

type employeeService struct {
	empRepo  domain.EmployeeRepository
	deptRepo domain.DepartmentRepository
}

func NewEmployeeService(
	empRepo domain.EmployeeRepository,
	deptRepo domain.DepartmentRepository,
) domain.EmployeeService {
	return &employeeService{
		empRepo:  empRepo,
		deptRepo: deptRepo,
	}
}

func (s *employeeService) Create(ctx context.Context, departmentID int, fullName, position string, hiredAt *time.Time) (*domain.Employee, error) {
	// Проверяем что подразделение существует
	if _, err := s.deptRepo.GetByID(ctx, departmentID); err != nil {
		return nil, fmt.Errorf("department: %w", domain.ErrNotFound)
	}

	fullName = strings.TrimSpace(fullName)
	if err := validateField("full_name", fullName); err != nil {
		return nil, err
	}

	position = strings.TrimSpace(position)
	if err := validateField("position", position); err != nil {
		return nil, err
	}

	e := &domain.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}
	if err := s.empRepo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func validateField(field, value string) error {
	if value == "" || len(value) > 200 {
		return fmt.Errorf("%s must be between 1 and 200 characters: %w", field, domain.ErrInvalidInput)
	}
	return nil
}
