package repository

import (
	"context"
	"errors"
	"log"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
}

type userRepository struct {
	db *gorm.DB
}

// CreateUserAccount implements [UserRepositoryInterface].
func (u *userRepository) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	modelRole := model.Role{}
	err := u.db.Where("name = ?", "Customer").First(&modelRole).Error

	if err != nil {
		log.Printf("[UserRepository-1] CreateUserAccount: %v", err)
		return err
	}

	modelUser := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Roles:    []model.Role{modelRole},
	}

	if err := u.db.Create(&modelUser).Error; err != nil {
		log.Printf("[UserRepository-2] CreateUserAccount: %v", err)
		return err
	}

	currentTime := time.Now()
	modelVerify := model.VerificationToken{
		UserID:    modelUser.ID,
		Token:     req.Token,
		TokenType: "email_verification",
		ExpiresAt: currentTime.Add(time.Hour * 1),
	}

	if err := u.db.Create(&modelVerify).Error; err != nil {
		log.Printf("[UserRepository-3] CreateUserAccount: %v", err)
		return err
	}

	return nil
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.Where("email = ? AND is_verified = ?", email, true).
		Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Println("[UserRepository-1] GetUserByEmail: User Not Found")
			return nil, err
		}
		log.Fatal("[UserRepository-1] GetUserByEmail: %v", err)
		return nil, err
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      email,
		Password:   modelUser.Password,
		RoleName:   modelUser.Roles[0].Name,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}
