package adapter

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"product-service/config"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/service"

	"github.com/labstack/echo/v4"
)

// ContextUser is the key the middleware stores the session under. Handlers read
// it back with c.Get(ContextUser).(string).
const ContextUser = "user"

// customerRole is the only role barred from /admin routes.
const customerRole = "Customer"

type MiddlewareAdapterInterface interface {
	CheckToken() echo.MiddlewareFunc
}

type middlewareAdapter struct {
	jwtService service.JWTServiceInterface
}

// CheckToken implements [MiddlewareAdapterInterface].
//
// Two independent checks: the JWT signature must hold, and the token must still
// have a live session in Redis. The second is what makes logout effective — a
// token that is valid until its exp claims nothing once user-service drops the
// session.
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			respErr := response.DefaultResponse{}

			authHeader := e.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("[MiddlewareAdapter-1] CheckToken: missing Authorization header")
				respErr.Message = "Missing Authorization header"
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			// "Bearer " is stripped only when present, so a bare token would
			// otherwise be accepted with its prefix intact and fail with a
			// confusing signature error instead of a clear 401.
			if tokenString == authHeader {
				log.Printf("[MiddlewareAdapter-2] CheckToken: malformed Authorization header")
				respErr.Message = "Authorization header must be in the form: Bearer <token>"
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			if _, err := m.jwtService.ValidateToken(tokenString); err != nil {
				log.Printf("[MiddlewareAdapter-3] CheckToken: %v", err)
				respErr.Message = "Invalid or expired token"
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			// Reached only for requests that carry a token, so unauthenticated
			// traffic never touches Redis.
			getSession, err := config.NewRedisClient().Get(e.Request().Context(), tokenString).Result()
			if err != nil || getSession == "" {
				log.Printf("[MiddlewareAdapter-4] CheckToken: no session for token: %v", err)
				respErr.Message = "Invalid or expired token"
				return e.JSON(http.StatusUnauthorized, respErr)
			}

			jwtUserData := entity.JwtUserData{}
			if err := json.Unmarshal([]byte(getSession), &jwtUserData); err != nil {
				// the session exists but is not the JSON we wrote: ours to fix,
				// not the client's, so 500 rather than 401
				log.Printf("[MiddlewareAdapter-5] CheckToken: %v", err)
				respErr.Message = "Internal server error"
				return e.JSON(http.StatusInternalServerError, respErr)
			}

			// /admin/... is the admin surface; a customer token must not reach it
			segments := strings.Split(strings.Trim(e.Request().URL.Path, "/"), "/")
			if jwtUserData.RoleName == customerRole && segments[0] == "admin" {
				log.Printf("[MiddlewareAdapter-6] CheckToken: role %q denied on /admin", jwtUserData.RoleName)
				respErr.Message = "Forbidden"
				return e.JSON(http.StatusForbidden, respErr)
			}

			// stored as the raw session string to stay compatible with the
			// handlers, which do c.Get(ContextUser).(string)
			e.Set(ContextUser, getSession)
			return next(e)
		}
	}
}

func NewMiddlewareAdapter(jwtService service.JWTServiceInterface) MiddlewareAdapterInterface {
	return &middlewareAdapter{jwtService: jwtService}
}
