package models

import "gorm.io/gorm"

// User represents a user entity for the database
type User struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `gorm:"uniqueIndex;not null" json:"username"`
	Email      string `gorm:"uniqueIndex;not null" json:"email"`
	Password   string `gorm:"not null" json:"-"` // Password should not be exposed
}

// PublicUser represents the user information safe to expose in APIs
type PublicUser struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}
