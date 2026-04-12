package models

import "time"

// User is the canonical user record used across handlers and tests.
type User struct {
	ID           string    `json:"id" gorm:"primaryKey;type:text"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Name         string    `json:"name" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	Role         string    `json:"role" gorm:"not null;default:user"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
