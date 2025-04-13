package models

import (
	"time"
)

type AuthMethod string

const (
	UsernamePassword AuthMethod = "username_password"
)

type Tenant struct {
	ID        string       `json:"id" gorm:"primaryKey"`
	Name      string       `json:"name" gorm:"not null"`
	Config    TenantConfig `json:"config" gorm:"foreignKey:TenantID"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type TenantConfig struct {
	ID              string     `json:"id" gorm:"primaryKey"`
	TenantID        string     `json:"tenant_id" gorm:"not null;uniqueIndex"`
	AuthMethod      AuthMethod `json:"auth_method" gorm:"not null"`
	JWTDuration     int        `json:"jwt_duration" gorm:"not null"`
	RateLimitIP     int        `json:"rate_limit_ip" gorm:"not null"`
	RateLimitUser   int        `json:"rate_limit_user" gorm:"not null"`
	RateLimitWindow int        `json:"rate_limit_window" gorm:"not null"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (c *TenantConfig) Update(authMethod AuthMethod, jwtDuration, rateLimitIP, rateLimitUser, rateLimitWindow int) {
	c.AuthMethod = authMethod
	c.JWTDuration = jwtDuration
	c.RateLimitIP = rateLimitIP
	c.RateLimitUser = rateLimitUser
	c.RateLimitWindow = rateLimitWindow
}

func DefaultConfig(tenantID string) *TenantConfig {
	return &TenantConfig{
		TenantID:        tenantID,
		AuthMethod:      UsernamePassword,
		JWTDuration:     60,
		RateLimitIP:     100,
		RateLimitUser:   50,
		RateLimitWindow: 60,
	}
}
