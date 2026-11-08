package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-service/config"
	"product-service/internal/adapter/handler"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/service"
	validators "product-service/utils/validator"

	"github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer() {
	cfg := config.NewConfig()
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[RunServer-1] %v", err)
		return
	}

	// Built eagerly so a missing REDIS_HOST/REDIS_PORT fails at startup instead
	// of surfacing as a misleading 401 from the token middleware.
	config.NewRedisClient()

	categoryRepo := repository.NewCategoryRepository(db.DB)
	categoryService := service.NewCategoryService(categoryRepo)

	jwtService := service.NewJwtService(cfg)

	e := echo.New()
	e.Use(middleware.CORS())

	customValidator := validators.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator

	handler.NewCategoryHandler(e, categoryService, jwtService)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err = e.Start(":" + cfg.App.AppPort)
		if err != nil {
			log.Fatalf("[RunServer-2] %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	<-quit

	log.Print("[RunServer-3] Shutting down server of 5 second...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	e.Shutdown(ctx)
}
