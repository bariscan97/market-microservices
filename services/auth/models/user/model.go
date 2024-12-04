package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID `gorm:"type:char(36);primary_key"`
	Username            string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Email               string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password            string    `gorm:"type:text;not null"`
	IsVerified          bool      `gorm:"default:false"`
	VerificationToken   string    `gorm:"type:text;default:NULL"`
	ResetPasswordToken  string    `gorm:"type:text;default:NULL"`
	ResetPasswordExpire time.Time `gorm:"default:NULL"`
	Role                string    `gorm:"type:varchar(20);default:'user';not null"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=5,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=24"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=24"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=8,max=24"`
}
