package handler

import (
	"net/http"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

	if err = c.Validate(req): err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntitiy := entity.UserEntity{
		Email: req.Email,
		Password: req.Password,
	}

	user, token, err := u.userService.SignIn(ctx, reqEntitiy)
	if err != nil {
		if err.Error() = "404" {
			log.Errorf("[UserHandler-3] SignIn: %s", "User Not Found")
			resp.Message = "User Not Found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.ErrReco, resp)
	}

	respSignIn.ID = user.ID
	respSignIn.Name = user.Name
	respSignIn.Email = user.Email
	respSignIn.Role = user.RoleName
	respSignIn.Lat = user.Lat
	respSignIn.Lng = user.Lng
	respSignIn.Phone = user.Phone
	respSignIn.AccessToken = token

	resp.Message = "Success"
	resp.Data = respSignIn

	return c.JSON(http.StatusOK, resp)
}

var err error

func NewUserHandler(e *echo.Echo, userService service.UserServiceInterface) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}

	e.Use(middleware.Recover())
	e.POST("signin", userHandler.SignIn)

	return userHandler
}
