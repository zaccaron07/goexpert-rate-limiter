package usecases

import (
	"context"
	"fmt"
	"time"

	appconfig "goexpert-rate-limiter/internal/application/config"
	"goexpert-rate-limiter/internal/domain/entities"
	"goexpert-rate-limiter/internal/domain/repositories"
)

type RateLimiterUseCase interface {
	CheckRateLimit(ctx context.Context, request entities.RateLimitRequest) (entities.RateLimitResult, error)
}

type rateLimiterUseCase struct {
	repository repositories.RateLimiterRepository
	config     *appconfig.RateLimiterConfig
}

func NewRateLimiterUseCase(repository repositories.RateLimiterRepository, config *appconfig.RateLimiterConfig) RateLimiterUseCase {
	return &rateLimiterUseCase{
		repository: repository,
		config:     config,
	}
}

func (s *rateLimiterUseCase) CheckRateLimit(ctx context.Context, request entities.RateLimitRequest) (entities.RateLimitResult, error) {
	isBlocked, blockUntil, err := s.repository.IsBlocked(ctx, request.Key, request.Type)
	if err != nil {
		return entities.RateLimitResult{}, fmt.Errorf("failed to check blocked status: %w", err)
	}

	if isBlocked {
		return entities.RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  blockUntil,
			BlockUntil: blockUntil,
		}, nil
	}

	var maxRequests int
	var blockDuration time.Duration

	switch request.Type {
	case entities.IPRateLimit:
		maxRequests = s.config.IPRequestsPerSecond
		blockDuration = time.Duration(s.config.IPBlockDurationSeconds) * time.Second
	case entities.TokenRateLimit:
		maxRequests = s.config.TokenRequestsPerSecond
		blockDuration = time.Duration(s.config.TokenBlockDurationSeconds) * time.Second
	default:
		return entities.RateLimitResult{}, fmt.Errorf("invalid rate limit type: %s", request.Type)
	}

	expiration := time.Duration(1) * time.Second
	newCount, err := s.repository.IncrementRequestCount(ctx, request.Key, request.Type, expiration)
	if err != nil {
		return entities.RateLimitResult{}, fmt.Errorf("failed to increment request count: %w", err)
	}

	if newCount > maxRequests {
		blockUntil := time.Now().Add(blockDuration)
		if err := s.repository.SetBlocked(ctx, request.Key, request.Type, blockUntil); err != nil {
			return entities.RateLimitResult{}, fmt.Errorf("failed to set blocked status: %w", err)
		}

		return entities.RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  blockUntil,
			BlockUntil: blockUntil,
		}, nil
	}

	remaining := maxRequests - newCount
	resetTime := time.Now().Add(expiration)

	return entities.RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		ResetTime:  resetTime,
		BlockUntil: time.Time{},
	}, nil
}
