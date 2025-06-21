package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username          string     `json:"username" gorm:"unique;not null"`
	Email             string     `json:"email" gorm:"unique;not null"`
	Password          string     `json:"-"`
	IsActive          bool       `json:"is_active" gorm:"default:true"`
	IsBanned          bool       `json:"is_banned" gorm:"default:false"`
	FailedAttempts    int        `json:"failed_attempts" gorm:"default:0"`
	LastFailedAttempt time.Time  `json:"last_failed_attempt,omitempty"`
	LockoutUntil      *time.Time `json:"lockout_until,omitempty"`
}

type Content struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      int    `json:"user_id"`
}

type ContentBody struct {
	ContentID uint   `bson:"content_id"`
	Body      string `bson:"body"`
}
