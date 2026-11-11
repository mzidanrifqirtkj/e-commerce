package model

import "time"

type Product struct {
	ID           int64  `gorm:"primaryKey"`
	CategorySlug string `gorm:"index"`
	ParentID     *int64
	Name         string
	Image        string
	Description  string
	RegularPrice float64
	SalePrice    float64 `gorm:"index"`
	Unit         string
	Weight       int
	Stock        int
	Variant      int
	Status       string    `gorm:"index"`
	Category     Category  `gorm:"foreignKey:CategorySlug;references:Slug"`
	Children     []Product `gorm:"foreignKey:ParentID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
