package repository

import (
	"context"
	"errors"
	"log"
	"time"
	"user-service/internal/core/domain/apperror"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"gorm.io/gorm"
)

type VerificationTokenRepositoryInterface interface {
	CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error
	GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error)
	DeleteTokenByID(ctx context.Context, id int64) error
}

type VerificationTokenRepository struct {
	db *gorm.DB
}

// GetDataByToken implements [VerificationTokenRepositoryInterface].
func (v *VerificationTokenRepository) GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error) {
	modelToken := model.VerificationToken{}
	if err := v.db.WithContext(ctx).Where("token = ?", token).First(&modelToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = apperror.ErrNotFound
			log.Printf("[VerificationTokenRepository-1] GetDataByToken: %v", err)
			return nil, err
		}
		log.Printf("[VerificationTokenRepository-2] GetDataByToken: %v", err)
		return nil, err
	}

	if time.Now().After(modelToken.ExpiresAt) {
		log.Printf("[VerificationTokenRepository-3] GetDataByToken: token expired")
		return nil, apperror.ErrUnauthorized
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
// The caller sets ExpiresAt; a zero value falls back to 24 hours.
func (v *VerificationTokenRepository) CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error {
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	modelVerificationToken := model.VerificationToken{
		UserID:    req.UserID,
		Token:     req.Token,
		TokenType: req.TokenType,
		ExpiresAt: expiresAt,
	}

	if err := v.db.WithContext(ctx).Create(&modelVerificationToken).Error; err != nil {
		log.Printf("[VerificationTokenRepository-1] CreateVerificationToken: %v", err)
		return err
	}

	return nil
}

// DeleteTokenByID implements [VerificationTokenRepositoryInterface].
func (v *VerificationTokenRepository) DeleteTokenByID(ctx context.Context, id int64) error {
	if err := v.db.WithContext(ctx).Where("id = ?", id).Delete(&model.VerificationToken{}).Error; err != nil {
		log.Printf("[VerificationTokenRepository-5] DeleteTokenByID: %v", err)
		return err
	}

	return nil
}

func NewVerificationTokenRepository(db *gorm.DB) VerificationTokenRepositoryInterface {
	return &VerificationTokenRepository{db: db}
}
