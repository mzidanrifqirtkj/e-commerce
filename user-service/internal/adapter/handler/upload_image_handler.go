package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/adapter/storage"
	"user-service/internal/core/service"
	"uuid"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type UploadImageInterface interface {
	UploadImage(c echo.Context) error
}

type uploadImage struct {
	storageHandler storage.SupabaseInterface
}

// UploadImage implements [UploadImageInterface].
func (u *uploadImage) UploadImage(c echo.Context) error {
	var resp = response.DefaultResponse{}
	file, err := c.FormFile("photo")
	if err != nil {
		log.Errorf("[UploadImage-1] Upload Image %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	src, err := file.Open()
	if err != nil {
		log.Errorf("[UploadImage-2] Upload Image %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	defer src.Close()

	fileBuffer := new(bytes.Buffer)
	_, err = io.Copy(fileBuffer, src)
	if err != nil {
		log.Errorf("[UploadImage-3] Upload Image %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	newFileName := fmt.Sprintf("%s_%d%s", uuid.New(), time.Now().Unix(), getExtension(file.Filename))

	uploadPath := fmt.Sprintf("public/uploads/%s", newFileName)
	url, err := u.storageHandler.UploadFile(uploadPath, file.Header.Get("Content-Type"), fileBuffer)
	if err != nil {
		log.Errorf("[UploadImage-4] Upload Image %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = map[string]string{"image_url": url}

	return c.JSON(http.StatusOK, resp)
}

func getExtension(fileName string) string {
	ext := "." + fileName[len(fileName)-3:]
	if len(fileName) > 4 && fileName[len(fileName)-4] == '.' {
		ext = "." + fileName[len(fileName)-4:]
	}
	return ext
}

func NewUploadImage(e *echo.Echo, userService service.UserServiceInterface, jwtService service.JWTServiceInterface, storageHandler storage.SupabaseInterface) UploadImageInterface {
	res := &uploadImage{
		storageHandler: storageHandler,
	}

	mid := adapter.NewMiddlewareAdapter(jwtService)
	e.POST("/auth/profile/image-upload", res.UploadImage, mid.CheckToken())
	return res
}
