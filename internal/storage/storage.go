package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tajious/heimdall/internal/config"
	"github.com/tajious/heimdall/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Storage interface {
	CreateTenant(ctx context.Context, tenant *models.Tenant) error
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
	UpdateTenantConfig(ctx context.Context, config *models.TenantConfig) error
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	UpdateUserLastLogin(ctx context.Context, userID string) error
	GetDB() *gorm.DB
	ListTenants(ctx context.Context, page, pageSize int) ([]*models.Tenant, int64, error)
}

type PostgresStorage struct {
	db *gorm.DB
}

type InMemoryStorage struct {
	tenants map[string]*models.Tenant
	users   map[string]*models.User
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Tenant{}, &models.TenantConfig{}, &models.User{}); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tenants: make(map[string]*models.Tenant),
		users:   make(map[string]*models.User),
	}
}

func (s *PostgresStorage) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	return s.db.WithContext(ctx).Create(tenant).Error
}

func (s *PostgresStorage) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.WithContext(ctx).Preload("Config").First(&tenant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (s *PostgresStorage) UpdateTenantConfig(ctx context.Context, config *models.TenantConfig) error {
	return s.db.WithContext(ctx).Save(config).Error
}

func (s *PostgresStorage) CreateUser(ctx context.Context, user *models.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *PostgresStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "phone = ?", phone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) UpdateUserLastLogin(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login", time.Now()).Error
}

func (s *PostgresStorage) GetDB() *gorm.DB {
	return s.db
}

func (s *PostgresStorage) ListTenants(ctx context.Context, page, pageSize int) ([]*models.Tenant, int64, error) {
	var tenants []*models.Tenant
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.WithContext(ctx).Model(&models.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).Preload("Config").Offset(offset).Limit(pageSize).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

func (s *InMemoryStorage) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	s.tenants[tenant.ID] = tenant
	return nil
}

func (s *InMemoryStorage) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	tenant, exists := s.tenants[id]
	if !exists {
		return nil, ErrTenantNotFound
	}
	return tenant, nil
}

func (s *InMemoryStorage) UpdateTenantConfig(ctx context.Context, config *models.TenantConfig) error {
	tenant, exists := s.tenants[config.TenantID]
	if !exists {
		return ErrTenantNotFound
	}
	tenant.Config = *config
	return nil
}

func (s *InMemoryStorage) CreateUser(ctx context.Context, user *models.User) error {
	s.users[user.ID] = user
	return nil
}

func (s *InMemoryStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *InMemoryStorage) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	for _, user := range s.users {
		if user.Phone == phone {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *InMemoryStorage) UpdateUserLastLogin(ctx context.Context, userID string) error {
	user, exists := s.users[userID]
	if !exists {
		return ErrUserNotFound
	}
	user.LastLogin = time.Now()
	return nil
}

func (s *InMemoryStorage) GetDB() *gorm.DB {
	return nil
}

func (s *InMemoryStorage) ListTenants(ctx context.Context, page, pageSize int) ([]*models.Tenant, int64, error) {
	var tenants []*models.Tenant
	total := int64(len(s.tenants))

	offset := (page - 1) * pageSize
	end := offset + pageSize
	if end > int(total) {
		end = int(total)
	}

	for _, tenant := range s.tenants {
		tenants = append(tenants, tenant)
	}

	if offset >= int(total) {
		return []*models.Tenant{}, total, nil
	}

	return tenants[offset:end], total, nil
}

func BuildDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
}
