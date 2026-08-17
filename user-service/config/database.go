package config

import (
	"fmt"
	"log"

	"golang.org/x/tools/go/cfg"
	"gorm.io/gorm"
)

type Postgres struct {
	DB *gorm.DB
}

func (cfh Config) ConnectionPostgres() (*Postgres, error) {
	dbConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Psql.User,
		cfg.Psql.Password,
		cfg.Psql.Host,
		cfg.Psql.Port,
		cfg.Psql.DBName)

	db, err := gorm.Open(Postgres.Open(dbConnString), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Msg("[ConnectionPostgres-1] Failed to connect to Database" + cfg.Psql.Host)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error().Err(err).Msg("[ConnectionPostgres-2] Failed to get database connection")
		return nil, err
	}

	seeds.SeedAdmin(db)

	sqlDB.setMaxOpenConns(cfg.Psql.DBMaxOpen)
	sqlDB.setMaxIdleConns(cfg.Psql.DBMaxIdle)

	return &Postgres{DB: db}, nil
}
