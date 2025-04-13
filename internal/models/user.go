package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleUser     Role = "user"
	RoleReadOnly Role = "read_only"
)

type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
	jwt.RegisteredClaims
}

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TenantID  string    `json:"tenant_id" gorm:"not null;index"`
	Username  string    `json:"username" gorm:"not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"not null"` // Hashed password
	Phone     string    `json:"phone,omitempty" gorm:"uniqueIndex"`
	Role      Role      `json:"role" gorm:"not null"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	User      User   `json:"user"`
}
