package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"
	"user-service/internal/core/domain/apperror"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	UpdateUserVerified(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error
	GetProfileUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error

	// Modul Customers Admin
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, int64, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// GetCustomerByID implements [UserRepositoryInterface].
func (u *userRepository) GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, int64, int64, error) {
	modelUser := model.User{}
	if err := u.db.Where("id = ?", customerID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[UserRepository-1] GetCustomerByID: Customer not found")
			return nil, 0, 0, apperror.ErrNotFound
		}
		log.Printf("[UserRepository-2] GetCustomerByID: %v", err)
		return nil, 0, 0, err
	}

	if modelUser.ID == 0 || len(modelUser.Roles) < 1 {
		log.Printf("[UserRepository-3] GetCustomerByID: Customer not found")
		return nil, 0, 0, apperror.ErrNotFound
	}

	return &entity.UserEntity{
		ID:      modelUser.ID,
		Name:    modelUser.Name,
		Email:   modelUser.Email,
		RoleID:  modelUser.Roles[0].ID,
		Phone:   modelUser.Phone,
		Address: modelUser.Address,
		Photo:   modelUser.Photo,
		Lat:     modelUser.Lat,
		Lng:     modelUser.Lng,
	}, 0, 0, nil
}

// GetAll implements [UserRepositoryInterface].
func (u *userRepository) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error) {
	modelUsers := []model.User{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := u.db.Preload("Roles", "name = ?", "Customer").
		Where("name ILIKE ? OR email ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")

	if err := sqlMain.Model(&modelUsers).Count(&countData).Error; err != nil {
		log.Printf("[UserRepository-1] GetCustomerAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))

	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelUsers).Error; err != nil {
		log.Printf("[UserRepository-2] GetCustomerAll: No Customer found")
		return nil, 0, 0, err
	}

	// if err := u.db.Preload("Roles", "name = ?", "Customer").Find(&modelUsers).Error; err {
	// 	log.Printf( "[UserRepository-1] GetCustomerAll: %v", err)
	// 	return nil, err
	// }

	if len(modelUsers) < 1 {
		err := apperror.ErrNotFound
		log.Printf("[UserRepository-3] GetCustomerAll: No Customer found")
		return nil, 0, 0, err
	}

	respEntities := []entity.UserEntity{}
	for _, val := range modelUsers {
		respEntities = append(respEntities, entity.UserEntity{
			ID:       val.ID,
			Name:     val.Name,
			Email:    val.Email,
			RoleName: val.Roles[0].Name,
			Phone:    val.Phone,
			Photo:    val.Photo,
		})
	}
	return respEntities, countData, int64(totalPage), nil
}

// UpdateDataUser implements [UserRepositoryInterface].
func (u *userRepository) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	modelUser := model.User{
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Phone:   req.Phone,
		Photo:   req.Photo,
		Lat:     req.Lat,
		Lng:     req.Lng,
	}

	if err := u.db.Where("id = ? AND is_verified = true", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = apperror.ErrNotFound
			log.Printf("[UserRepository-1] UpdateDataUser: %v", err)
			return err
		}
		log.Printf("[UserRepository-2] UpdateDataUser: %v", err)
		return err
	}

	if err := u.db.Save(&modelUser).Error; err != nil {
		log.Printf("[UserRepository-3] UpdateDataUser: %v", err)
		return err
	}

	return nil
}

// UpdatePasswordByID implements [UserRepositoryInterface].
func (u *userRepository) UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = apperror.ErrNotFound
			log.Printf("[UserRepository-1] UpdatePasswordByID: %v", err)
			return err
		}
		log.Printf("[UserRepository-2] UpdatePasswordByID: %v", err)
		return err
	}

	modelUser.Password = req.Password
	if err := u.db.WithContext(ctx).Save(&modelUser).Error; err != nil {
		log.Printf("[UserRepository-3] UpdatePasswordByID: %v", err)
		return err
	}

	return nil
}

// UpdateUserVerified implements [UserRepositoryInterface].
func (u *userRepository) UpdateUserVerified(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", userID).
		Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = apperror.ErrNotFound
			log.Printf("[UserRepository-1] UpdateUserVerified: %v", err)
			return nil, err
		}
		log.Printf("[UserRepository-2] UpdateUserVerified: %v", err)
		return nil, err
	}

	modelUser.IsVerified = true
	if err := u.db.WithContext(ctx).Save(&modelUser).Error; err != nil {
		log.Printf("[UserRepository-3] UpdateUserVerified: %v", err)
		return nil, err
	}

	return toUserEntity(modelUser), nil
}

// GetProfileUserByID implements [UserRepositoryInterface].
func (u *userRepository) GetProfileUserByID(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", userID).
		Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = apperror.ErrNotFound
			log.Printf("[UserRepository-1] GetProfileUserByID: %v", err)
			return nil, err
		}
		log.Printf("[UserRepository-2] GetProfileUserByID: %v", err)
		return nil, err
	}

	return toUserEntity(modelUser), nil
}

// toUserEntity maps a gorm user model to the entity the service layer expects.
func toUserEntity(modelUser model.User) *entity.UserEntity {
	roleName := ""
	if len(modelUser.Roles) > 0 {
		roleName = modelUser.Roles[0].Name
	}

	return &entity.UserEntity{
		ID:         modelUser.ID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		Password:   modelUser.Password,
		RoleName:   roleName,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}
}

// CreateUserAccount implements [UserRepositoryInterface].
func (u *userRepository) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	modelRole := model.Role{}
	err := u.db.WithContext(ctx).Where("name = ?", "Customer").First(&modelRole).Error

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

	if err := u.db.WithContext(ctx).Create(&modelUser).Error; err != nil {
		log.Printf("[UserRepository-2] CreateUserAccount: %v", err)
		return err
	}

	modelVerify := model.VerificationToken{
		UserID:    modelUser.ID,
		Token:     req.Token,
		TokenType: entity.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := u.db.WithContext(ctx).Create(&modelVerify).Error; err != nil {
		log.Printf("[UserRepository-3] CreateUserAccount: %v", err)
		return err
	}

	return nil
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	modelUser := model.User{}

	if err := u.db.WithContext(ctx).Where("email = ? AND is_verified = ?", email, true).
		Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = apperror.ErrNotFound
			log.Println("[UserRepository-1] GetUserByEmail: User Not Found")
			return nil, err
		}
		log.Printf("[UserRepository-1] GetUserByEmail: %v", err)
		return nil, err
	}

	return toUserEntity(modelUser), nil
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}
