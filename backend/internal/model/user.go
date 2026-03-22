package model

type User struct {
	Base
	Username     string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Role         string `gorm:"type:varchar(50);default:'teacher'" json:"role"`
	IsVerified   bool   `gorm:"default:false" json:"is_verified"`
}
