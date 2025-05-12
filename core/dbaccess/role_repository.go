package dbaccess

import (
	"context"

	"FinanceGolang/core/domain"

	"gorm.io/gorm"
	// "errors"
)

// RoleRepository интерфейс репозитория ролей
type RoleRepository interface {
	Repository[domain.Role]
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	GetByUserID(ctx context.Context, userID uint) ([]domain.Role, error)
	AddPermission(ctx context.Context, roleID uint, permission string) error
	RemovePermission(ctx context.Context, roleID uint, permission string) error
	GetByPermission(ctx context.Context, permission string) ([]domain.Role, error)
	GetActiveRoles(ctx context.Context) ([]domain.Role, error)
	UpdatePermissions(ctx context.Context, roleID uint, permissions []string) error
}

// roleRepository реализация репозитория ролей
type roleRepository struct {
	BaseRepository[domain.Role]
}

// RoleRepositoryInstance создает новый репозиторий ролей
func RoleRepositoryInstance(db *gorm.DB) RoleRepository {
	return &roleRepository{
		BaseRepository: *NewBaseRepository[domain.Role](db),
	}
}

// Create создает новую роль
func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := role.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Create(role).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// GetByID получает роль по ID
func (r *roleRepository) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &role, nil
}

// GetByName получает роль по имени
func (r *roleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &role, nil
}

// GetByUserID получает роли пользователя
func (r *roleRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return roles, nil
}

// GetByPermission получает роли по разрешению
func (r *roleRepository) GetByPermission(ctx context.Context, permission string) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.Where("permissions @> ?", permission).Find(&roles).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return roles, nil
}

// GetActiveRoles получает активные роли
func (r *roleRepository) GetActiveRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.Where("is_active = ?", true).Find(&roles).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return roles, nil
}

// Update обновляет роль
func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := role.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Save(role).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// AddPermission добавляет разрешение к роли
func (r *roleRepository) AddPermission(ctx context.Context, roleID uint, permission string) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Role{}).Where("id = ?", roleID).
			Update("permissions", gorm.Expr("array_append(permissions, ?)", permission)).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// RemovePermission удаляет разрешение из роли
func (r *roleRepository) RemovePermission(ctx context.Context, roleID uint, permission string) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Role{}).Where("id = ?", roleID).
			Update("permissions", gorm.Expr("array_remove(permissions, ?)", permission)).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// UpdatePermissions обновляет список разрешений роли
func (r *roleRepository) UpdatePermissions(ctx context.Context, roleID uint, permissions []string) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Role{}).Where("id = ?", roleID).
			Update("permissions", permissions).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// Delete удаляет роль
func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&domain.Role{}, id).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// List получает список ролей
func (r *roleRepository) List(ctx context.Context, offset, limit int) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return roles, nil
}

// Count возвращает количество ролей
func (r *roleRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.Role{}).Count(&count).Error; err != nil {
		return 0, r.HandleError(err)
	}
	return count, nil
}
