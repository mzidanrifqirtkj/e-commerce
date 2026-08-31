package service

import (
	"context"
	"errors"
	"log"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	"user-service/utils/conv"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
}

type userService struct {
	repo repository.UserRepositoryInterface
}

func (u *userService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Println("[UserService-1] SignIn: %v", err)
		return nil, "", err
	}

	if checkPass := conv.CheckPasswordHash(req.Password, user.Password); !checkPass {
		err = errors.New("Password is Incorrect")
		log.Fatal("[UserService-1] SignIn: %v", err)
		return nil, "", err
	}

	return user, "", err
}

func NewUserService(repo repository.UserRepositoryInterface) UserServiceInterface {
	return &userService{repo: repo}
}
