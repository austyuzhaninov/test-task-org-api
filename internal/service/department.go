package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

type departmentService struct {
	deptRepo domain.DepartmentRepository
	empRepo  domain.EmployeeRepository
}

func NewDepartmentService(
	deptRepo domain.DepartmentRepository,
	empRepo domain.EmployeeRepository,
) domain.DepartmentService {
	return &departmentService{
		deptRepo: deptRepo,
		empRepo:  empRepo,
	}
}

// ────────────────────────────────────────────────────────────────────────────

func (s *departmentService) Create(ctx context.Context, name string, parentID *int) (*domain.Department, error) {
	name = strings.TrimSpace(name)
	if err := validateName(name); err != nil {
		return nil, err
	}

	// Проверяем что parent существует
	if parentID != nil {
		if _, err := s.deptRepo.GetByID(ctx, *parentID); err != nil {
			return nil, fmt.Errorf("parent department: %w", domain.ErrNotFound)
		}
	}

	d := &domain.Department{
		Name:     name,
		ParentID: parentID,
	}
	if err := s.deptRepo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// ────────────────────────────────────────────────────────────────────────────

func (s *departmentService) GetByID(ctx context.Context, id, depth int, includeEmployees bool) (*domain.DepartmentNode, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}

	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	node := &domain.DepartmentNode{Department: dept}

	if includeEmployees {
		emps, err := s.empRepo.ListByDepartment(ctx, id)
		if err != nil {
			return nil, err
		}
		node.Employees = emps
	}

	// Рекурсивно строим дерево до depth
	children, err := s.buildChildren(ctx, id, depth-1, includeEmployees)
	if err != nil {
		return nil, err
	}
	node.Children = children

	return node, nil
}

// buildChildren рекурсивно загружает дочерние подразделения.
func (s *departmentService) buildChildren(ctx context.Context, parentID, depth int, includeEmployees bool) ([]*domain.DepartmentNode, error) {
	if depth < 0 {
		return nil, nil
	}

	children, err := s.deptRepo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	nodes := make([]*domain.DepartmentNode, 0, len(children))
	for _, child := range children {
		node := &domain.DepartmentNode{Department: child}

		if includeEmployees {
			emps, err := s.empRepo.ListByDepartment(ctx, child.ID)
			if err != nil {
				return nil, err
			}
			node.Employees = emps
		}

		grandChildren, err := s.buildChildren(ctx, child.ID, depth-1, includeEmployees)
		if err != nil {
			return nil, err
		}
		node.Children = grandChildren
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// ────────────────────────────────────────────────────────────────────────────

func (s *departmentService) Update(ctx context.Context, id int, name *string, parentID *int, clearParent bool) (*domain.Department, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if err := validateName(trimmed); err != nil {
			return nil, err
		}
		dept.Name = trimmed
	}

	switch {
	case clearParent:
		// Явный null в запросе — делаем подразделение корневым
		dept.ParentID = nil

	case parentID != nil:
		newParent := *parentID

		// Нельзя сделать родителем самого себя
		if newParent == id {
			return nil, fmt.Errorf("department cannot be its own parent: %w", domain.ErrConflict)
		}

		// Нельзя переместить в своё поддерево — проверяем через рекурсивный CTE
		inSubtree, err := s.deptRepo.ExistsInSubtree(ctx, id, newParent)
		if err != nil {
			return nil, err
		}
		if inSubtree {
			return nil, fmt.Errorf("cannot move department into its own subtree: %w", domain.ErrConflict)
		}

		// Проверяем что новый parent существует
		if _, err := s.deptRepo.GetByID(ctx, newParent); err != nil {
			return nil, fmt.Errorf("parent department: %w", domain.ErrNotFound)
		}

		dept.ParentID = parentID
	}

	if err := s.deptRepo.Update(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

// ────────────────────────────────────────────────────────────────────────────

func (s *departmentService) Delete(ctx context.Context, id int, mode string, reassignTo *int) error {
	if _, err := s.deptRepo.GetByID(ctx, id); err != nil {
		return err
	}

	switch mode {
	case "cascade":
		// ON DELETE CASCADE на уровне БД сам удалит сотрудников и дочерние отделы
		return s.deptRepo.Delete(ctx, id)

	case "reassign":
		if reassignTo == nil {
			return fmt.Errorf("reassign_to_department_id is required: %w", domain.ErrInvalidInput)
		}
		if _, err := s.deptRepo.GetByID(ctx, *reassignTo); err != nil {
			return fmt.Errorf("reassign target department: %w", domain.ErrNotFound)
		}

		// Защита — нельзя удалить отдел у которого есть дочерние
		children, err := s.deptRepo.GetChildren(ctx, id)
		if err != nil {
			return err
		}
		if len(children) > 0 {
			return fmt.Errorf("department has %d child departments, use cascade or move them first: %w",
				len(children), domain.ErrConflict)
		}

		return s.deptRepo.DeleteWithReassign(ctx, id, *reassignTo)

	default:
		return fmt.Errorf("unknown delete mode %q, use cascade or reassign: %w", mode, domain.ErrInvalidInput)
	}
}

// ────────────────────────────────────────────────────────────────────────────

func validateName(name string) error {
	if name == "" || len(name) > 200 {
		return fmt.Errorf("name must be between 1 and 200 characters: %w", domain.ErrInvalidInput)
	}
	return nil
}
