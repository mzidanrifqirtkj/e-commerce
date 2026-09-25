package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"user-service/config"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	"user-service/utils/conv"

	"github.com/google/uuid"
)

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	ForgotPassword(ctx context.Context, req entity.UserEntity) error
	VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error)
}

type userService struct {
	repo       repository.UserRepositoryInterface
	cfg        *config.Config
	jwtService JWTServiceInterface
	repoToken  repository.VerificationTokenRepositoryInterface
}

// VerifyToken implements [UserServiceInterface].
func (u *userService) VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error) {
	verifyToken, err := u.repoToken.GetDataByToken(ctx, token)
	if err != nil {
		log.Printf("[UserService-1] VerifyToken: %v", err)
		return nil, err
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

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      accessToken,
	}

	redisConn := config.NewRedisClient()
	err = redisConn.HSet(ctx, accessToken, sessionData).Err()
	if err != nil {
		log.Printf("[UserService-4] VerifyToken: %v", err)
		return nil, err
	}

	user.Token = accessToken
	return user, nil
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
		TokenType: "forgot_password",
	}

	err = u.repoToken.CreateVerificationToken(ctx, reqEntity)
	if err != nil {
		log.Printf("[UserService-2] ForgotPassword: %v", err)
		return err
	}

	urlForgot := fmt.Sprintf("%s/forgot-password?token=%s", u.cfg.App.UrlForgotPassword, token)
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

	urlVerify := fmt.Sprintf("%s/verify-account?token=%v", u.cfg.App.UrlForgotPassword, req.Token)
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
		err = errors.New("Password is Incorrect")
		log.Printf("[UserService-2] SignIn: %v", err)
		return nil, "", err
	}

	token, err := u.jwtService.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		log.Printf("[UserService-3] SignIn: %v", err)
		return nil, "", err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
	}

	redisConn := config.NewRedisClient()
	err = redisConn.HSet(ctx, token, sessionData).Err()
	if err != nil {
		log.Printf("[UserService-4] SignIn: %v", err)
		return nil, "", err
	}

	return user, token, err
}

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config, jwtService JWTServiceInterface, repoToken repository.VerificationTokenRepositoryInterface) UserServiceInterface {
	return &userService{repo: repo, cfg: cfg, jwtService: jwtService, repoToken: repoToken}
}
