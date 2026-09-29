package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"user-service/config"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/apperror"
	"user-service/internal/core/domain/entity"
	"user-service/utils/conv"

	"github.com/google/uuid"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	ForgotPassword(ctx context.Context, req entity.UserEntity) error
	VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error)
	UpdatePassword(ctx context.Context, req entity.UserEntity) error
	GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error)
}

type userService struct {
	repo       repository.UserRepositoryInterface
	cfg        *config.Config
	jwtService JWTServiceInterface
	repoToken  repository.VerificationTokenRepositoryInterface
}

// GetProfileUser implements [UserServiceInterface].
func (u *userService) GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	return u.repo.GetProfileUserByID(ctx, userID)
}

// UpdatePassword implements [UserServiceInterface].
func (u *userService) UpdatePassword(ctx context.Context, req entity.UserEntity) error {
	token, err := u.repoToken.GetDataByToken(ctx, req.Token)
	if err != nil {
		log.Printf("[UserService-1] UpdatePassword: %v", err)
		return err
	}

	if token.TokenType != "reset_password" {
		log.Printf("[UserService-2] UpdatePassword: wrong token type")
		return apperror.ErrUnauthorized
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Printf("[UserService-3] UpdatePassword: %v", err)
		return err
	}

	req.Password = password
	req.ID = token.UserID

	err = u.repo.UpdatePasswordByID(ctx, req)
	if err != nil {
		log.Printf("[UserService-4] UpdatePassword: %v", err)
		return err
	}

	if err = u.repoToken.DeleteTokenByID(ctx, token.ID); err != nil {
		log.Printf("[UserService-5] UpdatePassword: %v", err)
		return err
	}

	return nil
}

// VerifyToken implements [UserServiceInterface].
func (u *userService) VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error) {
	verifyToken, err := u.repoToken.GetDataByToken(ctx, token)
	if err != nil {
		log.Printf("[UserService-1] VerifyToken: %v", err)
		return nil, err
	}

	if verifyToken.TokenType != "email_verification" {
		log.Printf("[UserService-2] VerifyToken: wrong token type")
		return nil, apperror.ErrUnauthorized
	}

	user, err := u.repo.UpdateUserVerified(ctx, verifyToken.UserID)
	if err != nil {
		log.Printf("[UserService-2] VerifyToken: %v", err)
		return nil, err
	}

	accessToken, err := u.jwtService.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		log.Printf("[UserService-3] VerifyToken: %v", err)
		return nil, err
	}

	session, err := toSession(user, accessToken)
	if err != nil {
		log.Printf("[UserService-4] VerifyToken: %v", err)
		return nil, err
	}

	redisConn := config.NewRedisClient()
	if err = redisConn.Set(ctx, accessToken, session, time.Hour*23).Err(); err != nil {
		log.Printf("[UserService-5] VerifyToken: %v", err)
		return nil, err
	}

	user.Token = accessToken
	return user, nil
}

// toSession stores everything CheckToken needs in Redis as a JSON string.
// A map with mixed value types cannot be written by go-redis, only scalar or
// binary-marshaling values, so the session is serialized first.
func toSession(user *entity.UserEntity, accessToken string) (string, error) {
	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now(),
		"token":      accessToken,
	}

	raw, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ForgotPassword implements [UserServiceInterface].
func (u *userService) ForgotPassword(ctx context.Context, req entity.UserEntity) error {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("[UserService-1] ForgotPassword: %v", err)
		return err
	}

	token := uuid.New().String()
	reqEntity := entity.VerificationTokenEntity{
		UserID:    user.ID,
		Token:     token,
		TokenType: "reset_password",
	}

	err = u.repoToken.CreateVerificationToken(ctx, reqEntity)
	if err != nil {
		log.Printf("[UserService-2] ForgotPassword: %v", err)
		return err
	}

	urlForgot := fmt.Sprintf("%s?token=%s", u.cfg.App.UrlForgotPassword, token)
	messageParam := fmt.Sprintf("Please click link below for reset password: %v", urlForgot)
	err = message.PublishMessage(req.Email, messageParam, "forgot-password")
	if err != nil {
		log.Printf("[UserService-3] ForgotPassword: %v", err)
		return err
	}
	return nil
}

// CreateUserAccount implements [UserServiceInterface].
func (u *userService) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	password, err := conv.HashPassword(req.Password)

	if err != nil {
		log.Printf("[UserService-1] CreateUserAccount: %v", err)
		return err
	}

	req.Password = password
	token := uuid.New().String()
	req.Token = token

	err = u.repo.CreateUserAccount(ctx, req)
	if err != nil {
		log.Printf("[UserService-2] CreateUserAccount: %v", err)
		return err
	}

	urlVerify := fmt.Sprintf("%s/verify-account?token=%v", strings.TrimSuffix(u.cfg.App.UrlApp, "/"), req.Token)
	messageParam := fmt.Sprintf("Please verify your account with click link below: %v", urlVerify)
	err = message.PublishMessage(req.Email, messageParam, "email_verification")
	if err != nil {
		log.Printf("[UserService-3] CreateUserAccount: %v", err)
		return err
	}

	return nil
}

func (u *userService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("[UserService-1] SignIn: %v", err)
		return nil, "", err
	}

	if checkPass := conv.CheckPasswordHash(req.Password, user.Password); !checkPass {
		log.Printf("[UserService-2] SignIn: password is incorrect")
		return nil, "", apperror.ErrInvalidPassword
	}

	token, err := u.jwtService.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		log.Printf("[UserService-3] SignIn: %v", err)
		return nil, "", err
	}

	session, err := toSession(user, token)
	if err != nil {
		log.Printf("[UserService-4] SignIn: %v", err)
		return nil, "", err
	}

	redisConn := config.NewRedisClient()
	if err = redisConn.Set(ctx, token, session, time.Hour*23).Err(); err != nil {
		log.Printf("[UserService-5] SignIn: %v", err)
		return nil, "", err
	}

	return user, token, nil
}

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config, jwtService JWTServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) UserServiceInterface {
	return &userService{repo: repo, cfg: cfg, jwtService: jwtService, repoToken: repoToken}
}
