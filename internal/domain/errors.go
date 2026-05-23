package domain

import "errors"

var (
	// ErrNotFound — запись не найдена в БД.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict — нарушение бизнес-правила: дубль имени, цикл в дереве.
	ErrConflict = errors.New("conflict")

	// ErrInvalidInput — невалидные входные данные.
	ErrInvalidInput = errors.New("invalid input")
)
