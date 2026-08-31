package model

import "time"

type UserRole struct {
	ID        int64 `gorm:"primaryKey"`
	RoleID    int   `gorm:"index"`
	UserID    int   `gorm:"index"`
	Name      string
	Users     []User `gorm:"many2many:user_role"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (UserRole) TableName() string {
	return "user_role"
}
