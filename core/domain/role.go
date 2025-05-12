package domain

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrInvalidRoleName = errors.New("invalid role name")
	ErrRoleExists      = errors.New("role already exists")
)

const (
	RoleAdmin    = "ADMIN"
	RoleUser     = "USER"
	RoleManager  = "MANAGER"
	RoleOperator = "OPERATOR"
)

type Role struct {
	gorm.Model
	Name        string   `json:"name" gorm:"unique;not null" validate:"required,min=3,max=50"`
	Description string   `json:"description" gorm:"type:text"`
	Permissions []string `json:"permissions" gorm:"type:jsonb"`
	IsActive    bool     `json:"is_active" gorm:"default:true"`
	CreatedBy   uint     `json:"created_by"`
	UpdatedBy   uint     `json:"updated_by"`
	Users       []User   `json:"users" gorm:"many2many:user_roles;"`
}

// Validate проверяет все поля роли
func (r *Role) Validate() error {
	if err := r.ValidateName(); err != nil {
		return err
	}
	return nil
}

// ValidateName проверяет корректность имени роли
func (r *Role) ValidateName() error {
	switch r.Name {
	case RoleAdmin, RoleUser, RoleManager, RoleOperator:
		return nil
	default:
		return ErrInvalidRoleName
	}
}

// HasPermission проверяет наличие разрешения у роли
func (r *Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// AddPermission добавляет разрешение к роли
func (r *Role) AddPermission(permission string) {
	if !r.HasPermission(permission) {
		r.Permissions = append(r.Permissions, permission)
	}
}

// RemovePermission удаляет разрешение из роли
func (r *Role) RemovePermission(permission string) {
	for i, p := range r.Permissions {
		if p == permission {
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			break
		}
	}
}

// BeforeCreate хук для валидации перед созданием
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	return r.Validate()
}

// BeforeUpdate хук для валидации перед обновлением
func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	return r.Validate()
}

// ToDTO преобразует модель в DTO
func (r *Role) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":          r.ID,
		"name":        r.Name,
		"description": r.Description,
		"permissions": r.Permissions,
		"is_active":   r.IsActive,
		"created_by":  r.CreatedBy,
		"updated_by":  r.UpdatedBy,
		"created_at":  r.CreatedAt,
		"updated_at":  r.UpdatedAt,
	}
}

// GetDefaultRoles возвращает список ролей по умолчанию
func GetDefaultRoles() []Role {
	return []Role{
		{
			Name:        RoleAdmin,
			Description: "Администратор системы",
			Permissions: []string{
				"user:create",
				"user:read",
				"user:update",
				"user:delete",
				"role:create",
				"role:read",
				"role:update",
				"role:delete",
				"account:create",
				"account:read",
				"account:update",
				"account:delete",
				"card:create",
				"card:read",
				"card:update",
				"card:delete",
				"credit:create",
				"credit:read",
				"credit:update",
				"credit:delete",
			},
		},
		{
			Name:        RoleUser,
			Description: "Обычный пользователь",
			Permissions: []string{
				"account:read",
				"card:read",
				"credit:read",
			},
		},
		{
			Name:        RoleManager,
			Description: "Менеджер",
			Permissions: []string{
				"user:read",
				"account:read",
				"account:update",
				"card:read",
				"card:update",
				"credit:read",
				"credit:update",
			},
		},
		{
			Name:        RoleOperator,
			Description: "Оператор",
			Permissions: []string{
				"user:read",
				"account:read",
				"card:read",
				"credit:read",
			},
		},
	}
}

type UserRole struct {
	UserID uint `json:"user_id" gorm:"primaryKey"`
	RoleID uint `json:"role_id" gorm:"primaryKey"`
}
