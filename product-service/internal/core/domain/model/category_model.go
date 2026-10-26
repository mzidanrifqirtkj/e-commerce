package model

import "time"

type Category struct {
	ID          int64 `gorm:"primaryKey"`
	ParentID    *int64
	Name        string
	Icon        string
	Status      bool   `gorm:"index"`
	Slug        string `gorm:"index"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Products    []Product `gorm:"foreignKey:CategorySlug;references:Slug"`
}
