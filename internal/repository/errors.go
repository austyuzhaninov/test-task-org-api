package repository

import (
	"errors"
	"fmt"

	"github.com/austyuzhaninov/test-task-org-api/internal/domain"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// mapDBError переводит технические ошибки ORM/БД в чистые доменные ошибки.
// Это централизованное место: если завтра изменится ORM или СУБД,
// мы поменяем логику только здесь.
func mapDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("name already exists in this parent: %w", domain.ErrConflict)
		case "23503": // foreign_key_violation
			return fmt.Errorf("related entity not found: %w", domain.ErrNotFound)
		}
	}

	return err
}
