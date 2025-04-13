package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tajious/heimdall/internal/models"
	"github.com/tajious/heimdall/internal/storage"
	"github.com/tajious/heimdall/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	storage     storage.Storage
	jwtSecret   string
	jwtDuration time.Duration
}

func NewAuthHandler(storage storage.Storage, jwtSecret string, jwtDuration time.Duration) *AuthHandler {
	return &AuthHandler{
		storage:     storage,
		jwtSecret:   jwtSecret,
		jwtDuration: jwtDuration,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request using shared validator
	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get the tenant from the request context
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
		})
	}

	// Get tenant configuration
	tenant, err := h.storage.GetTenant(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid tenant",
		})
	}

	// Handle authentication
	user, authErr := h.authenticateWithUsernamePassword(c.Context(), req)
	if authErr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Verify user belongs to the tenant
	if user.TenantID != tenantID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid tenant",
		})
	}

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Update last login time
	if err := h.storage.UpdateUserLastLogin(c.Context(), user.ID); err != nil {
		// Log the error but don't fail the request
		c.Locals("error", err)
	}

	return c.JSON(models.LoginResponse{
		Token:     token,
		ExpiresIn: int(tenant.Config.JWTDuration),
		User:      *user,
	})
}

func (h *AuthHandler) authenticateWithUsernamePassword(ctx context.Context, req models.LoginRequest) (*models.User, error) {
	if req.Username == "" || req.Password == "" {
		return nil, storage.ErrInvalidCredentials
	}

	user, err := h.storage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, storage.ErrInvalidCredentials
	}

	return user, nil
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := models.Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.jwtDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}

	// Extract token from "Bearer <token>"
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	// Get user from storage
	user, err := h.storage.GetUserByUsername(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get tenant configuration
	tenant, err := h.storage.GetTenant(c.Context(), claims.TenantID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid tenant",
		})
	}

	return c.JSON(fiber.Map{
		"valid": true,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
		"tenant": fiber.Map{
			"id":     tenant.ID,
			"name":   tenant.Name,
			"config": tenant.Config,
		},
		"expires_at": claims.ExpiresAt,
	})
}

// ListUsersRequest represents the query parameters for listing users
type ListUsersRequest struct {
	Page     int    `query:"page" validate:"min=1"`
	PageSize int    `query:"page_size" validate:"min=1,max=100"`
	Search   string `query:"search"`
	Role     string `query:"role"`
	SortBy   string `query:"sort_by" validate:"oneof=username role created_at last_login"`
	SortDir  string `query:"sort_dir" validate:"oneof=asc desc"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []models.User `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// ListUsers handles listing users with pagination, search, filtering, and sorting
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	// Get tenant ID from path parameter
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
		})
	}

	// Check if tenant exists
	if _, err := h.storage.GetTenant(c.Context(), tenantID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tenant not found",
		})
	}

	// Get user's tenant ID from context (set by auth middleware)
	userTenantID := c.Locals("tenant_id").(string)
	if userTenantID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User tenant ID not found",
		})
	}

	// Verify user has access to the requested tenant
	if userTenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied to this tenant",
		})
	}

	// Parse and validate query parameters
	var req ListUsersRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters",
		})
	}

	// Set default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	// Validate request
	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Build query
	query := h.storage.GetDB().Model(&models.User{}).Where("tenant_id = ?", tenantID)

	// Apply search
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("username LIKE ? OR phone LIKE ?", searchPattern, searchPattern)
	}

	// Apply role filter
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count users",
		})
	}

	// Calculate total pages
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	// Apply sorting
	sortField := req.SortBy
	if sortField == "created_at" {
		sortField = "created_at"
	} else if sortField == "last_login" {
		sortField = "last_login"
	}
	query = query.Order(sortField + " " + req.SortDir)

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// Execute query
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	// Return response
	return c.JSON(ListUsersResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}
