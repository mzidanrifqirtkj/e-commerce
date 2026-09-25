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

type VerificationTokenRepositoryInterface interface {
	CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error
	GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error)
}

type VerificationTokenRepository struct {
	db *gorm.DB
}

// GetDataByToken implements [VerificationTokenRepositoryInterface].
func (v *VerificationTokenRepository) GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error) {
	modelToken := model.VerificationToken{}
	if err := v.db.Where("token = ?", token).First(&modelToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Printf("[VerificationTokenRepository-1] GetDataByToken: %v", err)
			return nil, err
		}
		log.Printf("[VerificationTokenRepository-2] GetDataByToken: %v", err)
		return nil, err
	}

	currentTime := time.Now()
	if currentTime.After(modelToken.ExpiresAt) {
		err := errors.New("401")
		log.Printf("[VerificationTokenRepository-3] CreateVerificationToken: %v", err)
		return nil, err
	}

	return &entity.VerificationTokenEntity{
		ID:        int64(modelToken.ID),
		UserID:    modelToken.UserID,
		Token:     token,
		TokenType: modelToken.TokenType,
		ExpiresAt: modelToken.ExpiresAt,
	}, nil

}

// CreateVerificationToken implements [VerificationTokenRepositoryInterface].
func (v *VerificationTokenRepository) CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error {
	modelVerificationToken := model.VerificationToken{
		UserID:    req.UserID,
		Token:     req.Token,
		TokenType: req.TokenType,
	}

	if err := v.db.Create(&modelVerificationToken).Error; err != nil {
		log.Printf("[VerificationTokenRepository-1] CreateVerificationToken: %v", err)
		return err
	}

	return nil
}

func NewVerificationTokenRepository(db *gorm.DB) VerificationTokenRepositoryInterface {
	return &VerificationTokenRepository{db: db}
}
