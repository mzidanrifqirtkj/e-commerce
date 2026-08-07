# User Service

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- golang-migrate

## Setup

### 1. Install migration tool

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Add to PATH (add to `~/.zshrc` or `~/.bashrc`):

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### 2. Start infrastructure

```bash
docker compose -f internal/docker-compose.yml up -d
```

Services:
- Postgres: `localhost:5432` (db: `ecommerce`, user: `postgres`, pass: `postgres`)
- RabbitMQ: `localhost:5672` / Management UI: `localhost:15672`
- Elasticsearch: `localhost:9200`
- Redis: `localhost:6379`

### 3. Create migration

```bash
migrate create -ext sql -dir database/migrations -seq create_users_table
```

### 4. Run migrations

```bash
migrate -path database/migrations -database "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable" up
```

### 5. Rollback migration

```bash
migrate -path database/migrations -database "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable" down 1
```