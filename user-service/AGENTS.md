# User Service - Project Context & Learning Guide

## Tech Stack
- **Language**: Go 1.25.7
- **Web Framework**: Echo v4 (`github.com/labstack/echo/v4`)
- **Database & ORM**: PostgreSQL with GORM (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- **Migrations**: Goose / Golang-migrate
- **CLI & Config**: Cobra, Viper (`github.com/spf13/cobra`, `github.com/spf13/viper`)
- **Authentication**: JWT (`github.com/dgrijalva/jwt-go`)
- **Caching / Broker**: Redis, RabbitMQ

## Architecture
Clean Architecture / Layered structure under `internal/`:
- `core/domain/`: Entities and models (data structures)
- `core/service/`: Business logic (e.g., `user_service.go`, `jwt_service.go`)
- `adapter/handler/`: HTTP handlers / controllers using Echo (`user_handler.go`)
- `adapter/repository/`: Database queries (`user_repository.go`)
- `cmd/`: CLI commands via Cobra (`root.go`, `start.go`)
- `config/`: Configuration loaders (Database, Redis, RabbitMQ)

## Communication Style for Learner
- Clear, simple Indonesian (or English matching user's prompt).
- Explain *why* and *how* step by step using code snippets from the repo.
- Keep answers structured and easy to digest for beginners.
