package app

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/echo/v5/middleware"
)

func RunServer() {
	e := echo.New()
	e.Use(middleware.CORS())

	customValidator :=
}
