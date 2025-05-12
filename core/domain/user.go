package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidPassword = errors.New("invalid password format")
	ErrInvalidFIO      = errors.New("invalid FIO format")
	ErrEmailExists     = errors.New("email already exists")
	ErrUsernameExists  = errors.New("username already exists")
	ErrWrongPassword   = errors.New("wrong password")
)

type User struct {
	gorm.Model
	Fio       string    `json:"fio" gorm:"not null" validate:"required,min=3,max=100"`
	Username  string    `json:"username" gorm:"unique;not null" validate:"required,min=3,max=50,alphanum"`
	Email     string    `json:"email" gorm:"unique;not null" validate:"required,email"`
	Password  string    `json:"password" gorm:"not null" validate:"required,min=8"`
	Roles     []Role    `json:"roles" gorm:"many2many:user_roles;"`
	LastLogin time.Time `json:"last_login"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
}

// Validate проверяет все поля пользователя
func (u *User) Validate() error {
	if err := u.ValidateEmail(); err != nil {
		return err
	}
	if err := u.ValidateUsername(); err != nil {
		return err
	}
	if err := u.ValidatePassword(); err != nil {
		return err
	}
	if err := u.ValidateFIO(); err != nil {
		return err
	}
	return nil
}

// ValidateEmail проверяет корректность email
func (u *User) ValidateEmail() error {
	emailRegex := regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateUsername проверяет корректность username
func (u *User) ValidateUsername() error {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	if !usernameRegex.MatchString(u.Username) {
		return ErrInvalidUsername
	}
	return nil
}

// ValidatePassword проверяет корректность пароля
func (u *User) ValidatePassword() error {
	if len(u.Password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

// ValidateFIO проверяет корректность ФИО
func (u *User) ValidateFIO() error {
	fio := strings.TrimSpace(u.Fio)
	if len(fio) < 3 || len(fio) > 100 {
		return ErrInvalidFIO
	}
	return nil
}

// HashPassword хеширует пароль пользователя
func (u *User) HashPassword() error {
	if u.Password == "" {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// BeforeCreate хук для проверки уникальности и хеширования пароля перед созданием
func (u *User) BeforeCreate(tx *gorm.DB) error {
	var count int64

	// Проверка уникальности email
	tx.Model(&User{}).Where("email = ?", u.Email).Count(&count)
	if count > 0 {
		return ErrEmailExists
	}

	// Проверка уникальности username
	tx.Model(&User{}).Where("username = ?", u.Username).Count(&count)
	if count > 0 {
		return ErrUsernameExists
	}

	// Хеширование пароля
	return u.HashPassword()
}

// BeforeUpdate хук для проверки уникальности и хеширования пароля перед обновлением
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	var count int64

	// Проверка уникальности email
	tx.Model(&User{}).Where("email = ? AND id != ?", u.Email, u.ID).Count(&count)
	if count > 0 {
		return ErrEmailExists
	}

	// Проверка уникальности username
	tx.Model(&User{}).Where("username = ? AND id != ?", u.Username, u.ID).Count(&count)
	if count > 0 {
		return ErrUsernameExists
	}

	// Хеширование пароля только если он был изменен
	if u.Password != "" {
		return u.HashPassword()
	}
	return nil
}

func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// ToDTO преобразует модель в DTO
func (u *User) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"fio":        u.Fio,
		"username":   u.Username,
		"email":      u.Email,
		"roles":      u.Roles,
		"last_login": u.LastLogin,
		"is_active":  u.IsActive,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}

// CheckPassword проверяет соответствие пароля хешу
func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return ErrWrongPassword
	}
	return nil
}
