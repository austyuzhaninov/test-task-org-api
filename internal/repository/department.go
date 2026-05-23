package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

// departmentModel — GORM-модель, изолирована внутри репозитория.
// domain.Department не знает ни о GORM, ни о тегах json.
type departmentModel struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null"`
	ParentID  *int
	CreatedAt gormTime `gorm:"autoCreateTime"`
}

func (departmentModel) TableName() string { return "departments" }

// gormTime нужен чтобы не тащить time.Time в теги — используем стандартный тип.
// Оставляем просто time.Time через алиас для читаемости модели.
type gormTime = interface{}

// ────────────────────────────────────────────────────────────────────────────

type DepartmentRepo struct {
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) *DepartmentRepo {
	return &DepartmentRepo{db: db}
}

// toModel конвертирует доменную сущность в GORM-модель.
func toDepModel(d *domain.Department) *departmentModel {
	return &departmentModel{
		ID:       d.ID,
		Name:     d.Name,
		ParentID: d.ParentID,
	}
}

// toDomain конвертирует GORM-модель в доменную сущность.
func toDepDomain(m *departmentModel) *domain.Department {
	return &domain.Department{
		ID:       m.ID,
		Name:     m.Name,
		ParentID: m.ParentID,
	}
}

// ────────────────────────────────────────────────────────────────────────────

func (r *DepartmentRepo) Create(ctx context.Context, d *domain.Department) error {
	m := toDepModel(d)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	d.ID = m.ID
	return nil
}

func (r *DepartmentRepo) GetByID(ctx context.Context, id int) (*domain.Department, error) {
	var m departmentModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDepDomain(&m), nil
}

func (r *DepartmentRepo) GetChildren(ctx context.Context, parentID int) ([]*domain.Department, error) {
	var models []departmentModel
	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("created_at").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domain.Department, len(models))
	for i := range models {
		result[i] = toDepDomain(&models[i])
	}
	return result, nil
}

func (r *DepartmentRepo) Update(ctx context.Context, d *domain.Department) error {
	// Updates с map гарантирует что NULL (ParentID = nil) тоже запишется.
	updates := map[string]any{
		"name":      d.Name,
		"parent_id": d.ParentID, // nil → SQL NULL
	}
	return r.db.WithContext(ctx).
		Model(&departmentModel{}).
		Where("id = ?", d.ID).
		Updates(updates).Error
}

func (r *DepartmentRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).
		Delete(&departmentModel{}, id).Error
}

// ExistsInSubtree рекурсивно проверяет, входит ли targetID в поддерево rootID.
// Используется при PATCH чтобы не допустить цикл в дереве.
func (r *DepartmentRepo) ExistsInSubtree(ctx context.Context, rootID, targetID int) (bool, error) {
	// Рекурсивный CTE — один запрос вместо N обходов в Go.
	query := `
		WITH RECURSIVE subtree AS (
			SELECT id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id FROM departments d
			INNER JOIN subtree s ON d.parent_id = s.id
		)
		SELECT EXISTS (SELECT 1 FROM subtree WHERE id = ?)
	`
	var exists bool
	if err := r.db.WithContext(ctx).Raw(query, rootID, targetID).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

// ReassignEmployees переводит всех сотрудников из fromDeptID в toDeptID.
func (r *DepartmentRepo) ReassignEmployees(ctx context.Context, fromDeptID, toDeptID int) error {
	return r.db.WithContext(ctx).
		Exec("UPDATE employees SET department_id = ? WHERE department_id = ?", toDeptID, fromDeptID).
		Error
}
