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

type CreateTenantRequest struct {
	Name            string            `json:"name" validate:"required,min=3,max=50"`
	Description     string            `json:"description" validate:"max=500"`
	AuthMethod      models.AuthMethod `json:"auth_method" validate:"required,oneof=username_password"`
	JWTDuration     int               `json:"jwt_duration" validate:"required,min=1"`
	RateLimitIP     int               `json:"rate_limit_ip" validate:"required,min=1"`
	RateLimitUser   int               `json:"rate_limit_user" validate:"required,min=1"`
	RateLimitWindow int               `json:"rate_limit_window" validate:"required,min=1"`
}

func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	var req CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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

	if err := h.storage.CreateTenant(c.Context(), tenant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create tenant",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tenant)
}

type UpdateTenantConfigRequest struct {
	AuthMethod      models.AuthMethod `json:"auth_method" validate:"required,oneof=username_password"`
	JWTDuration     int               `json:"jwt_duration" validate:"required,min=1"`
	RateLimitIP     int               `json:"rate_limit_ip" validate:"required,min=1"`
	RateLimitUser   int               `json:"rate_limit_user" validate:"required,min=1"`
	RateLimitWindow int               `json:"rate_limit_window" validate:"required,min=1"`
}

func (h *TenantHandler) UpdateTenantConfig(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
		})
	}

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

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	tenant.Config.AuthMethod = req.AuthMethod
	tenant.Config.JWTDuration = req.JWTDuration
	tenant.Config.RateLimitIP = req.RateLimitIP
	tenant.Config.RateLimitUser = req.RateLimitUser
	tenant.Config.RateLimitWindow = req.RateLimitWindow
	tenant.Config.UpdatedAt = time.Now()

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

type ListTenantsRequest struct {
	Page     int `query:"page" validate:"min=1"`
	PageSize int `query:"page_size" validate:"min=1,max=100"`
}

type ListTenantsResponse struct {
	Tenants    []*models.Tenant `json:"tenants"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

func (h *TenantHandler) ListTenants(c *fiber.Ctx) error {
	var req ListTenantsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters",
		})
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	tenants, total, err := h.storage.ListTenants(c.Context(), req.Page, req.PageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tenants",
		})
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return c.JSON(ListTenantsResponse{
		Tenants:    tenants,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

func (h *TenantHandler) GetTenant(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
		})
	}

	tenant, err := h.storage.GetTenant(c.Context(), tenantID)
	if err != nil {
		if err == storage.ErrTenantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Tenant not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tenant",
		})
	}

	return c.JSON(tenant)
}
