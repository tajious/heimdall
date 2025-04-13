package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tajious/heimdall/internal/models"
	"github.com/tajious/heimdall/internal/storage"
	"github.com/tajious/heimdall/internal/validation"
)

type TenantHandler struct {
	storage storage.Storage
}

func NewTenantHandler(storage storage.Storage) *TenantHandler {
	return &TenantHandler{
		storage: storage,
	}
}

// CreateTenantRequest represents the request body for tenant creation
type CreateTenantRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"max=500"`
	// Tenant configuration
	AuthMethod      models.AuthMethod `json:"auth_method" validate:"required,oneof=username_password"`
	JWTDuration     int               `json:"jwt_duration" validate:"required,min=1"`
	RateLimitIP     int               `json:"rate_limit_ip" validate:"required,min=1"`
	RateLimitUser   int               `json:"rate_limit_user" validate:"required,min=1"`
	RateLimitWindow int               `json:"rate_limit_window" validate:"required,min=1"`
}

// CreateTenant creates a new tenant with its configuration
func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	var req CreateTenantRequest
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

	// Create tenant
	tenant := &models.Tenant{
		Name: req.Name,
		Config: models.TenantConfig{
			AuthMethod:      req.AuthMethod,
			JWTDuration:     req.JWTDuration,
			RateLimitIP:     req.RateLimitIP,
			RateLimitUser:   req.RateLimitUser,
			RateLimitWindow: req.RateLimitWindow,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Save tenant
	if err := h.storage.CreateTenant(c.Context(), tenant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create tenant",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tenant)
}

// UpdateTenantConfigRequest represents the request body for updating tenant configuration
type UpdateTenantConfigRequest struct {
	AuthMethod      models.AuthMethod `json:"auth_method" validate:"required,oneof=username_password"`
	JWTDuration     int               `json:"jwt_duration" validate:"required,min=1"`
	RateLimitIP     int               `json:"rate_limit_ip" validate:"required,min=1"`
	RateLimitUser   int               `json:"rate_limit_user" validate:"required,min=1"`
	RateLimitWindow int               `json:"rate_limit_window" validate:"required,min=1"`
}

// UpdateTenantConfig updates the configuration for a tenant
func (h *TenantHandler) UpdateTenantConfig(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
		})
	}

	// Get the tenant to ensure it exists
	tenant, err := h.storage.GetTenant(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tenant not found",
		})
	}

	var req UpdateTenantConfigRequest
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

	// Update tenant configuration
	tenant.Config.AuthMethod = req.AuthMethod
	tenant.Config.JWTDuration = req.JWTDuration
	tenant.Config.RateLimitIP = req.RateLimitIP
	tenant.Config.RateLimitUser = req.RateLimitUser
	tenant.Config.RateLimitWindow = req.RateLimitWindow
	tenant.Config.UpdatedAt = time.Now()

	// Save updated configuration
	if err := h.storage.UpdateTenantConfig(c.Context(), &tenant.Config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update tenant configuration",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Tenant configuration updated successfully",
		"config":  tenant.Config,
	})
}
