package service

import (
	"time"
	"user-service/config"

	"github.com/dgrijalva/jwt-go"
)

type JWTServiceInterface interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtService struct {
	secretKey string
	issuer    string
}

// GenerateToken implements [JWTServiceInterface].
func (j *jwtService) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"iss":     j.issuer,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken implements [JWTServiceInterface].
func (j *jwtService) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
}

func NewJwtService(cfg *config.Config) JWTServiceInterface {
	return &jwtService{secretKey: cfg.App.JwtSecretkey, issuer: cfg.App.JwtIssuer}
}
