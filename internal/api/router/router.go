package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tajious/heimdall/internal/api/handlers"
	"github.com/tajious/heimdall/internal/middleware"
)

type Router struct {
	app            *fiber.App
	authHandler    *handlers.AuthHandler
	tenantHandler  *handlers.TenantHandler
	authMiddleware *middleware.AuthMiddleware
	rateLimiter    *middleware.RateLimiter
}

func NewRouter(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	tenantHandler *handlers.TenantHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
) *Router {
	return &Router{
		app:            app,
		authHandler:    authHandler,
		tenantHandler:  tenantHandler,
		authMiddleware: authMiddleware,
		rateLimiter:    rateLimiter,
	}
}

func (r *Router) SetupRoutes() {
	// Public routes
	r.app.Post("/api/v1/tenants", r.tenantHandler.CreateTenant)
	r.app.Post("/api/v1/:tenant_id/login", r.rateLimiter.RateLimit(middleware.RateLimitConfig{
		Enabled: true,
		Limit:   5,
		Window:  time.Minute,
	}), r.authHandler.Login)
	r.app.Post("/api/v1/validate-token", r.authHandler.ValidateToken)

	// Protected routes
	protected := r.app.Group("/api/v1", r.authMiddleware.Authenticate())
	protected.Get("/me", func(c *fiber.Ctx) error {
		user := c.Locals("user")
		return c.JSON(user)
	})
	protected.Put("/tenants/:tenant_id/config", r.tenantHandler.UpdateTenantConfig)
	protected.Get("/tenants/:tenant_id/users", r.authHandler.ListUsers)
}
