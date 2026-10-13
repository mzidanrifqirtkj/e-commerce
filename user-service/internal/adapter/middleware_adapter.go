package adapter

import (
	"encoding/json"
	"net/http"
	"strings"
	"user-service/config"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareAdapterInterface interface {
	CheckToken() echo.MiddlewareFunc
}

type middlewareAdapter struct {
	jwtService service.JWTServiceInterface
}

// CheckToken implements [MiddlewareAdapterInterface].
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			respErr := response.DefaultResponse{}
			authHeader := e.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: %s", "Missing Authorization header")
				respErr.Message = "Missing Authorization header"
				respErr.Data = nil
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			_, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", err.Error())
				respErr.Message = err.Error()
				respErr.Data = nil
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			// Reached only for requests that carry a token, so unauthenticated
			// traffic never touches Redis.
			redisConn := config.NewRedisClient()
			getSession, err := redisConn.Get(e.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-3] CheckToken: %s", "Invalid or expired token")
				respErr.Message = "Invalid or expired token"
				respErr.Data = nil
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			jwtUserData := entity.JwtUserData{}
			err = json.Unmarshal([]byte(getSession), &jwtUserData)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-4] CheckToken: %s", "Missing Authorization header")
				respErr.Message = err.Error()
				respErr.Data = nil
				return e.JSON(http.StatusInternalServerError, respErr)
			}

			path := e.Request().URL.Path
			segments := strings.Split(strings.Trim(path, "/"), "/")
			if jwtUserData.RoleName == "Customer" && segments[0] == "admin" {
				log.Errorf("[MiddlewareAdapter-4] CheckToken: %s", "customer cannot access admin routee")
				respErr.Message = "customer cannot adcces admin roeutes"
				respErr.Data = nil
				return e.JSON(http.StatusForbidden, respErr)
			}

			e.Set("user", getSession)
			return next(e)
		}
	}
}

func NewMiddlewareAdapter(jwtService service.JWTServiceInterface) MiddlewareAdapterInterface {
	return &middlewareAdapter{jwtService: jwtService}
}
