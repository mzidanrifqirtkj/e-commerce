package repository

import (
	"context"
	"log"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"gorm.io/gorm"
)

type VerificationTokenRepositoryInterface interface {
	CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error
}

type VerificationTokenRepository struct {
	db *gorm.DB
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
