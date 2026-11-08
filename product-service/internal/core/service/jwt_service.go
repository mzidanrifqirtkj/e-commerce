package service

import (
	"fmt"
	"log"

	"product-service/config"

	"github.com/golang-jwt/jwt"
)

type JWTServiceInterface interface {
	ValidateToken(encodedToken string) (*jwt.Token, error)
}

type jwtService struct {
	secretKey string
}

// ValidateToken implements [JWTServiceInterface].
//
// Signature and expiry only: the token is minted by user-service, so the
// session it points at lives in Redis and the middleware decides whether that
// session still exists.
func (j *jwtService) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(t *jwt.Token) (interface{}, error) {
		// Reject anything that is not HMAC so a token cannot claim another
		// signing algorithm and be verified with a key we did not intend.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(j.secretKey), nil
	})
}

func NewJwtService(cfg *config.Config) JWTServiceInterface {
	// Without the secret every token fails its signature, so every request gets
	// a 401 that reads as "wrong password" instead of "misconfigured deploy".
	// The key must match user-service's, since that is what signs the tokens.
	if cfg.App.JwtSecretkey == "" {
		log.Println("[JwtService-1] NewJwtService: JWT_SECRET_KEY is not set")
	}

	return &jwtService{secretKey: cfg.App.JwtSecretkey}
}
