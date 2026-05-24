package testhelper

import (
	"context"
	"time"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

// ─── Department Repository Mock ───────────────────────────────────────────────

type DeptRepoMock struct {
	Departments map[int]*domain.Department
	nextID      int

	// Хуки — позволяют переопределить поведение в конкретном тесте
	ExistsInSubtreeFn func(ctx context.Context, rootID, targetID int) (bool, error)
}

func NewDeptRepoMock() *DeptRepoMock {
	return &DeptRepoMock{
		Departments: make(map[int]*domain.Department),
		nextID:      1,
	}
}

func (m *DeptRepoMock) Create(ctx context.Context, d *domain.Department) error {
	d.ID = m.nextID
	d.CreatedAt = time.Now()
	m.nextID++
	m.Departments[d.ID] = d
	return nil
}

func (m *DeptRepoMock) GetByID(ctx context.Context, id int) (*domain.Department, error) {
	d, ok := m.Departments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return d, nil
}

func (m *DeptRepoMock) GetChildren(ctx context.Context, parentID int) ([]*domain.Department, error) {
	var result []*domain.Department
	for _, d := range m.Departments {
		if d.ParentID != nil && *d.ParentID == parentID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *DeptRepoMock) Update(ctx context.Context, d *domain.Department) error {
	if _, ok := m.Departments[d.ID]; !ok {
		return domain.ErrNotFound
	}
	m.Departments[d.ID] = d
	return nil
}

func (m *DeptRepoMock) Delete(ctx context.Context, id int) error {
	if _, ok := m.Departments[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.Departments, id)
	return nil
}

func (m *DeptRepoMock) ExistsInSubtree(ctx context.Context, rootID, targetID int) (bool, error) {
	if m.ExistsInSubtreeFn != nil {
		return m.ExistsInSubtreeFn(ctx, rootID, targetID)
	}
	return false, nil
}

// ─── Employee Repository Mock ─────────────────────────────────────────────────

type EmpRepoMock struct {
	Employees map[int]*domain.Employee
	nextID    int
}

func NewEmpRepoMock() *EmpRepoMock {
	return &EmpRepoMock{
		Employees: make(map[int]*domain.Employee),
		nextID:    1,
	}
}

func (m *EmpRepoMock) Create(ctx context.Context, e *domain.Employee) error {
	e.ID = m.nextID
	e.CreatedAt = time.Now()
	m.nextID++
	m.Employees[e.ID] = e
	return nil
}

func (m *EmpRepoMock) ListByDepartment(ctx context.Context, departmentID int) ([]*domain.Employee, error) {
	var result []*domain.Employee
	for _, e := range m.Employees {
		if e.DepartmentID == departmentID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *DeptRepoMock) DeleteWithReassign(ctx context.Context, id, reassignTo int) error {
	if _, ok := m.Departments[id]; !ok {
		return domain.ErrNotFound
	}
	// В моке просто удаляем — транзакция не нужна
	delete(m.Departments, id)
	return nil
}
