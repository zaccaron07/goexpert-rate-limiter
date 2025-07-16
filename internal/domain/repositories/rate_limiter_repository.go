package repositories

import (
	"context"
	"time"

	"goexpert-rate-limiter/internal/domain/entities"
)

type RateLimiterRepository interface {
	GetRequestCount(ctx context.Context, key string, rateLimitType entities.RateLimitType) (int, error)

	IncrementRequestCount(ctx context.Context, key string, rateLimitType entities.RateLimitType, expiration time.Duration) (int, error)

	SetBlocked(ctx context.Context, key string, rateLimitType entities.RateLimitType, blockUntil time.Time) error

	IsBlocked(ctx context.Context, key string, rateLimitType entities.RateLimitType) (bool, time.Time, error)

	GetBlockUntil(ctx context.Context, key string, rateLimitType entities.RateLimitType) (time.Time, error)

	Close() error
}
