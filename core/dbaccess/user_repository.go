package dbaccess

import (
	"context"

	"FinanceGolang/core/domain"
	"FinanceGolang/core/payloads"

	"gorm.io/gorm"
)

// UserRepository интерфейс репозитория пользователей
type UserRepository interface {
	Repository[domain.User]
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetWithRoles(ctx context.Context, id uint) (*domain.User, error)
	UpdatePassword(ctx context.Context, id uint, password string) error
	AddRole(ctx context.Context, userID, roleID uint) error
	RemoveRole(ctx context.Context, userID, roleID uint) error
	GetByRole(ctx context.Context, roleName string, offset, limit int) ([]domain.User, error)
}

// userRepository реализация репозитория пользователей
type userRepository struct {
	*BaseRepository[domain.User]
}

// UserRepositoryInstance создает новый репозиторий пользователей
func UserRepositoryInstance(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[domain.User](db),
	}
}

// Create создает нового пользователя
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := user.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Create(user).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// GetByID получает пользователя по ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &user, nil
}

// GetByUsername получает пользователя по username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.HandleError(err)
	}
	return &user, nil
}

// GetWithRoles получает пользователя с ролями
func (r *userRepository) GetWithRoles(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Roles").First(&user, id).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &user, nil
}

// Update обновляет пользователя
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := user.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Save(user).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// UpdatePassword обновляет пароль пользователя
func (r *userRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&domain.User{}).Where("id = ?", id).Update("password", password).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// Delete удаляет пользователя
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&domain.User{}, id).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// List получает список пользователей
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return users, nil
}

// Count возвращает количество пользователей
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return 0, r.HandleError(err)
	}
	return count, nil
}

// AddRole добавляет роль пользователю
func (r *userRepository) AddRole(ctx context.Context, userID, roleID uint) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		userRole := domain.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// RemoveRole удаляет роль у пользователя
func (r *userRepository) RemoveRole(ctx context.Context, userID, roleID uint) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&domain.UserRole{}).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// GetByRole получает пользователей по роли
func (r *userRepository) GetByRole(ctx context.Context, roleName string, offset, limit int) ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", roleName).
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return users, nil
}

// FindUserByUsername получает пользователя по username (устаревший метод)
func (r *userRepository) FindUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByUsernameWithoutPassword получает пользователя по username без пароля
func (r *userRepository) FindUserByUsernameWithoutPassword(username string) (*payloads.UserResponse, error) {
	var user domain.User
	if err := r.db.Select("id, username, email, created_at").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payloads.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
