package handler

import (
	"net/http"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/service"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

type UserHandlerInterface interface {
	SignIn(ctx echo.Context) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

var err error

// SignIn implements [UserHandlerInterface].
func (u *userHandler) SignIn(ctx echo.Context) error {
	var (
		req        = request.SigInRequest{}
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	panic("unimplemented")
}

var err error

func NewUserHandler(userService service.UserServiceInterface) UserHandlerInterface {
	return &userHandler{userService: userService}
}
