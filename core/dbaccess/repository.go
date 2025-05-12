package dbaccess

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
	ErrInvalidData   = errors.New("invalid data")
	ErrDatabaseError = errors.New("dbcore error")
)

// Repository базовый интерфейс для всех репозиториев
type Repository[T any] interface {
	// Create создает новую запись
	Create(ctx context.Context, entity *T) error
	// GetByID получает запись по ID
	GetByID(ctx context.Context, id uint) (*T, error)
	// Update обновляет запись
	Update(ctx context.Context, entity *T) error
	// Delete удаляет запись
	Delete(ctx context.Context, id uint) error
	// List получает список записей с пагинацией
	List(ctx context.Context, offset, limit int) ([]T, error)
	// Count возвращает общее количество записей
	Count(ctx context.Context) (int64, error)
}

// BaseRepository базовая реализация репозитория
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository создает новый базовый репозиторий
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// WithTransaction выполняет операции в транзакции
func (r *BaseRepository[T]) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return ErrDatabaseError
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return ErrDatabaseError
	}

	return nil
}

// HandleError обрабатывает ошибки базы данных
func (r *BaseRepository[T]) HandleError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrAlreadyExists
	default:
		return ErrDatabaseError
	}
}
