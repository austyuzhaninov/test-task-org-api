package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
)

// departmentModel — GORM-модель, изолирована внутри репозитория.
type departmentModel struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null"`
	ParentID  *int
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (departmentModel) TableName() string { return "departments" }

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
		ID:        d.ID,
		Name:      d.Name,
		ParentID:  d.ParentID,
		CreatedAt: d.CreatedAt,
	}
}

// toDepDomain конвертирует GORM-модель в доменную сущность.
func toDepDomain(m *departmentModel) *domain.Department {
	return &domain.Department{
		ID:        m.ID,
		Name:      m.Name,
		ParentID:  m.ParentID,
		CreatedAt: m.CreatedAt,
	}
}

// ────────────────────────────────────────────────────────────────────────────

func (r *DepartmentRepo) Create(ctx context.Context, d *domain.Department) error {
	m := toDepModel(d)

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapDBError(err) // Использование общей функции маппинга
	}

	d.ID = m.ID
	d.CreatedAt = m.CreatedAt
	return nil
}

func (r *DepartmentRepo) GetByID(ctx context.Context, id int) (*domain.Department, error) {
	var m departmentModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, mapDBError(err) // Передаем ошибку (включая ErrRecordNotFound) в маппер
	}

	return toDepDomain(&m), nil
}

func (r *DepartmentRepo) GetChildren(ctx context.Context, parentID int) ([]*domain.Department, error) {
	var models []departmentModel
	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("created_at").
		Find(&models).Error; err != nil {
		return nil, mapDBError(err)
	}

	result := make([]*domain.Department, len(models))
	for i := range models {
		result[i] = toDepDomain(&models[i])
	}
	return result, nil
}

func (r *DepartmentRepo) Update(ctx context.Context, d *domain.Department) error {
	updates := map[string]any{
		"name":      d.Name,
		"parent_id": d.ParentID,
	}

	result := r.db.WithContext(ctx).
		Model(&departmentModel{}).
		Where("id = ?", d.ID).
		Updates(updates)

	if result.Error != nil {
		return mapDBError(result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *DepartmentRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&departmentModel{})
	if result.Error != nil {
		return mapDBError(result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DepartmentRepo) DeleteWithReassign(ctx context.Context, id, reassignTo int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"UPDATE employees SET department_id = ? WHERE department_id = ?",
			reassignTo, id,
		).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", id).Delete(&departmentModel{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		return nil
	})
}

func (r *DepartmentRepo) ExistsInSubtree(ctx context.Context, rootID, targetID int) (bool, error) {
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
		return false, mapDBError(err)
	}
	return exists, nil
}
