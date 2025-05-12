package services

import (
	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/domain"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetUserByID(id uint) (*domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(id uint) error
}

type userService struct {
	userRepo dbaccess.UserRepository
}

func UserServiceInstance(userRepo dbaccess.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	return s.userRepo.GetByID(context.Background(), id)
}

func (s *userService) UpdateUser(user *domain.User) error {
	existingUser, err := s.userRepo.GetByID(context.Background(), user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// Сохраняем хеш пароля, если он не изменился
	if user.Password == "" {
		user.Password = existingUser.Password
	} else {
		// Хэшируем новый пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	return s.userRepo.Update(context.Background(), user)
}

func (s *userService) DeleteUser(id uint) error {
	existingUser, err := s.userRepo.GetByID(context.Background(), id)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}
	return s.userRepo.Delete(context.Background(), id)
}
