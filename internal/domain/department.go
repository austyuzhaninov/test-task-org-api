package domain

import (
	"context"
	"time"
)

// Department — чистая бизнес-сущность, без тегов GORM и JSON.
type Department struct {
	ID        int
	Name      string
	ParentID  *int
	CreatedAt time.Time
}

// DepartmentRepository — интерфейс принадлежит домену (используется сервисом).
// Реализация живёт в internal/repository.
type DepartmentRepository interface {
	Create(ctx context.Context, d *Department) error
	GetByID(ctx context.Context, id int) (*Department, error)
	GetChildren(ctx context.Context, parentID int) ([]*Department, error)
	Update(ctx context.Context, d *Department) error
	Delete(ctx context.Context, id int) error

	// ExistsInSubtree проверяет, входит ли targetID в поддерево rootID.
	// Используется для защиты от циклов при PATCH.
	ExistsInSubtree(ctx context.Context, rootID, targetID int) (bool, error)

	// DeleteWithReassign атомарно переводит сотрудников и удаляет отдел.
	DeleteWithReassign(ctx context.Context, id, reassignTo int) error
}

// DepartmentService — интерфейс бизнес-логики, используется хендлером.
type DepartmentService interface {
	Create(ctx context.Context, name string, parentID *int) (*Department, error)
	GetByID(ctx context.Context, id, depth int, includeEmployees bool) (*DepartmentNode, error)
	Update(ctx context.Context, id int, name *string, parentID *int, clearParent bool) (*Department, error)
	Delete(ctx context.Context, id int, mode string, reassignTo *int) error
}

// DepartmentNode — дерево подразделений для GET /departments/{id}.
// Живёт в домене т.к. это бизнес-структура, а не HTTP-ответ.
type DepartmentNode struct {
	*Department
	Employees []*Employee
	Children  []*DepartmentNode
}
