package domain

import "errors"

var (
	// ErrNotFound — сущность не найдена в БД.
	ErrNotFound = errors.New("not found")

	// ErrConflict — нарушение бизнес-правила (цикл, дубль имени и т.д.).
	ErrConflict = errors.New("conflict")

	// ErrInvalidInput — невалидные входные данные.
	ErrInvalidInput = errors.New("invalid input")
)
