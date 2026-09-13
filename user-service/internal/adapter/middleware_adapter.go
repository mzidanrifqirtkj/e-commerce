package adapter

import (
	"net/http"
	"strings"
	"user-service/config"
	"user-service/internal/adapter/handler/response"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareAdapterInterface interface {
	CheckToken() echo.MiddlewareFunc
}

type middlewareAdapter struct {
	cfg *config.Config
}

// CheckToken implements [MiddlewareAdapterInterface].
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			respErr := response.DefaultResponse{}
			redisConn := config.NewRedisClient()
			authHeader := e.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: %s", "Missing Authorization header")
				respErr.Message = "Missing Authorization header"
				respErr.Data = (*interface{})(nil)
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			getSession, err := redisConn.HGetAll(e.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", "Invalid or expired token")
				respErr.Message = err.Error()
				respErr.Data = (*interface{})(nil)
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			e.Set("user", getSession)
			return next(e)
		}
	}
}

func NewMiddlewareAdapter(cfg *config.Config) MiddlewareAdapterInterface {
	return &middlewareAdapter{cfg: cfg}
}
