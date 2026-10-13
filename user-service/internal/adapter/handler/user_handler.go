package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/apperror"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type UserHandlerInterface interface {
	SignIn(c echo.Context) error
	CreateUserAccount(c echo.Context) error
	ForgotPassword(c echo.Context) error
	VerifyAccount(c echo.Context) error
	UpdatePassword(c echo.Context) error
	GetProfileUser(c echo.Context) error
	UpdateDataUser(c echo.Context) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

// UpdateDataUser implements [UserHandlerInterface].
func (u *userHandler) UpdateDataUser(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		req         = request.UpdateDataUserRequest{}
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] UpdateDataUser %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}
	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] UpdateDataUser %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-3] UpdateDataUser %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-4] UpdateDataUser %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	latString := strconv.FormatFloat(req.Lat, 'g', -1, 64)
	lngString := strconv.FormatFloat(req.Lng, 'g', -1, 64)
	phoneString := fmt.Sprintf("%d", req.Phone)

	reqEntity := entity.UserEntity{
		ID:      userID,
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Phone:   phoneString,
		Photo:   req.Photo,
		Lat:     latString,
		Lng:     lngString,
	}

	err = u.userService.UpdateDataUser(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdateDataUser %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			resp.Message = "user not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil

	return c.JSON(http.StatusOK, resp)
}

// GetProfileUser implements [UserHandlerInterface].
func (u *userHandler) GetProfileUser(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		respProfile = response.ProfileResponse{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user, ok := c.Get("user").(string)
	if !ok || user == "" {
		log.Errorf("[UserHandler-1] GetProfileUser %s", "data token not found")
		resp.Message = "user not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}
	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] GetProfileUser %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID

	dataUser, err := u.userService.GetProfileUser(ctx, userID)
	if err != nil {
		log.Errorf("[UserHandler-3] GetProfileUser %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			resp.Message = "user not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respProfile.Address = dataUser.Address
	respProfile.Name = dataUser.Name
	respProfile.Email = dataUser.Email
	respProfile.ID = dataUser.ID
	respProfile.Lat = dataUser.Lat
	respProfile.Lng = dataUser.Lng
	respProfile.Phone = dataUser.Phone
	respProfile.Photo = dataUser.Photo
	respProfile.RoleName = dataUser.RoleName

	resp.Message = "success"
	resp.Data = respProfile

	return c.JSON(http.StatusOK, resp)
}

// UpdatePassword implements [UserHandlerInterface].
func (u *userHandler) UpdatePassword(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		req  = request.UpdatePasswordRequest{}
		ctx  = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Errorf("[UserHandler-1] UpdatePassword %s", "missing or invalid token")
		resp.Message = "missing or invalid token"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-2] UpdatePassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] UpdatePassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Errorf("[UserHandler-4] UpdatePassword %s", "new password and confirm password does not match")
		resp.Message = "new password and confirm password does not match"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	err := u.userService.UpdatePassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdatePassword %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		if errors.Is(err, apperror.ErrUnauthorized) {
			resp.Message = "Token expired or invalid"
			resp.Data = nil
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Data = nil
	resp.Message = "Password updated successfully"
	return c.JSON(http.StatusOK, resp)

}

// VerifyAccount implements [UserHandlerInterface].
func (u *userHandler) VerifyAccount(c echo.Context) error {
	var (
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Errorf("[UserHandler-1] VerifyAccount %s", "missing or invalid token")
		resp.Message = "missing or invalid token"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	user, err := u.userService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Errorf("[UserHandler-2] VerifyAccount: %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		if errors.Is(err, apperror.ErrUnauthorized) {
			resp.Message = "Token expired or invalid"
			resp.Data = nil
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respSignIn.ID = user.ID
	respSignIn.Name = user.Name
	respSignIn.Email = user.Email
	respSignIn.Role = user.RoleName
	respSignIn.Lat = user.Lat
	respSignIn.Lng = user.Lng
	respSignIn.Phone = user.Phone
	respSignIn.AccessToken = user.Token

	resp.Message = "Success"
	resp.Data = respSignIn

	return c.JSON(http.StatusOK, resp)
}

// ForgotPassword implements [UserHandlerInterface].
func (u *userHandler) ForgotPassword(c echo.Context) error {
	var (
		req  = request.ForgotPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-1] Forgot Password: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] Forgot Password: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email: req.Email,
	}

	err = u.userService.ForgotPassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-3] Forgot Password: %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)

}

func (u *userHandler) CreateUserAccount(c echo.Context) error {
	var (
		req  = request.SignUpRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-1] Create UserAccount: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] Create UserAccount: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if req.Password != req.PasswordConfirmation {
		log.Errorf("[UserHandler-3] Create UserAccount: %s", "Password Confirmation Not Match")
		resp.Message = "Password Confirmation Not Match"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err = u.userService.CreateUserAccount(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-4] Create UserAccount: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = nil
	return c.JSON(http.StatusCreated, resp)
}

// SignIn implements [UserHandlerInterface].
func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req        = request.SigInRequest{}
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.userService.SignIn(ctx, reqEntity)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			log.Errorf("[UserHandler-3] SignIn: %s", "User Not Found")
			resp.Message = "User Not Found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
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

func NewUserHandler(e *echo.Echo, userService service.UserServiceInterface, jwtService service.JWTServiceInterface) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}

	e.Use(middleware.Recover())
	e.POST("/signin", userHandler.SignIn)
	e.POST("/signup", userHandler.CreateUserAccount)
	e.POST("/forgot-password", userHandler.ForgotPassword)
	e.GET("/verify-account", userHandler.VerifyAccount)
	e.PUT("/update-password", userHandler.UpdatePassword)

	mid := adapter.NewMiddlewareAdapter(jwtService)

	adminGroup := e.Group("/admin")
	adminGroup.Use(mid.CheckToken())
	adminGroup.GET("/check", func(c echo.Context) error {
		return c.String(200, "Admin route accessed successfully")
	})

	profileGroup := e.Group("/profile")
	profileGroup.Use(mid.CheckToken())
	profileGroup.GET("/profile", userHandler.GetProfileUser)
	profileGroup.PUT("/profile", userHandler.UpdateDataUser)

	return userHandler
}
