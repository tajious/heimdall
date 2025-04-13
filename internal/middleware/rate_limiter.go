package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tajious/heimdall/internal/models"
)

// RateLimitStore defines the interface for rate limit storage
type RateLimitStore interface {
	Increment(ctx context.Context, key string, window time.Duration) (int, error)
	GetCount(ctx context.Context, key string) (int, error)
}

// RedisStore implements RateLimitStore using Redis
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
	// Use Redis pipeline for atomic operations
	pipe := s.client.Pipeline()

	// Increment the counter
	incr := pipe.Incr(ctx, key)

	// Set expiration if this is the first request
	pipe.Expire(ctx, key, window)

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return int(incr.Val()), nil
}

func (s *RedisStore) GetCount(ctx context.Context, key string) (int, error) {
	count, err := s.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// MemoryStore implements RateLimitStore using in-memory storage
type MemoryStore struct {
	mu    sync.RWMutex
	store map[string]*RateLimitEntry
}

type RateLimitEntry struct {
	Count     int
	ExpiresAt time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[string]*RateLimitEntry),
	}
}

func (s *MemoryStore) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean up expired entries
	now := time.Now()
	for k, entry := range s.store {
		if now.After(entry.ExpiresAt) {
			delete(s.store, k)
		}
	}

	// Get or create entry
	entry, exists := s.store[key]
	if !exists {
		entry = &RateLimitEntry{
			Count:     0,
			ExpiresAt: now.Add(window),
		}
		s.store[key] = entry
	}

	// Increment count
	entry.Count++
	return entry.Count, nil
}

func (s *MemoryStore) GetCount(ctx context.Context, key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.store[key]
	if !exists {
		return 0, nil
	}

	if time.Now().After(entry.ExpiresAt) {
		return 0, nil
	}

	return entry.Count, nil
}

// RateLimiter uses a RateLimitStore to enforce rate limits
type RateLimiter struct {
	store   RateLimitStore
	enabled bool
}

type RateLimitConfig struct {
	Enabled bool
	Limit   int
	Window  time.Duration
}

func NewRateLimiter(store RateLimitStore, enabled bool) *RateLimiter {
	return &RateLimiter{
		store:   store,
		enabled: enabled,
	}
}

func (r *RateLimiter) RateLimit(config RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !r.enabled || !config.Enabled {
			return c.Next()
		}

		// Get the IP address
		ip := c.IP()
		if ip == "" {
			ip = c.Context().RemoteIP().String()
		}

		// Get the user ID if available
		userID := ""
		if claims, ok := c.Locals("user").(*models.Claims); ok {
			userID = claims.UserID
		}

		// Create rate limit keys
		ipKey := fmt.Sprintf("rate_limit:ip:%s", ip)
		userKey := fmt.Sprintf("rate_limit:user:%s", userID)

		// Check IP rate limit
		if err := r.checkRateLimit(c.Context(), ipKey, config); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests from this IP",
			})
		}

		// Check user rate limit if user is authenticated
		if userID != "" {
			if err := r.checkRateLimit(c.Context(), userKey, config); err != nil {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Too many requests from this user",
				})
			}
		}

		return c.Next()
	}
}

func (r *RateLimiter) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) error {
	count, err := r.store.GetCount(ctx, key)
	if err != nil {
		return err
	}

	if count >= config.Limit {
		return fmt.Errorf("rate limit exceeded")
	}

	_, err = r.store.Increment(ctx, key, config.Window)
	return err
}
