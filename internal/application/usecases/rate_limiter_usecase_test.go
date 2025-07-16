package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appconfig "goexpert-rate-limiter/internal/application/config"
	"goexpert-rate-limiter/internal/domain/entities"
	"goexpert-rate-limiter/internal/domain/repositories"
)

type mockRepository struct {
	counts map[string]int
	blocks map[string]time.Time
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		counts: make(map[string]int),
		blocks: make(map[string]time.Time),
	}
}

func (m *mockRepository) GetRequestCount(ctx context.Context, key string, _ entities.RateLimitType) (int, error) {
	return m.counts[key], nil
}

func (m *mockRepository) IncrementRequestCount(ctx context.Context, key string, _ entities.RateLimitType, _ time.Duration) (int, error) {
	m.counts[key]++
	return m.counts[key], nil
}

func (m *mockRepository) SetBlocked(ctx context.Context, key string, _ entities.RateLimitType, until time.Time) error {
	m.blocks[key] = until
	return nil
}

func (m *mockRepository) IsBlocked(ctx context.Context, key string, _ entities.RateLimitType) (bool, time.Time, error) {
	until, ok := m.blocks[key]
	if !ok {
		return false, time.Time{}, nil
	}
	return time.Now().Before(until), until, nil
}

func (m *mockRepository) GetBlockUntil(ctx context.Context, key string, _ entities.RateLimitType) (time.Time, error) {
	return m.blocks[key], nil
}

func (m *mockRepository) Close() error { return nil }

var _ repositories.RateLimiterRepository = (*mockRepository)(nil)

func TestRateLimiter_Allowed(t *testing.T) {
	repo := newMockRepository()
	cfg := appconfig.RateLimiterConfig{
		IPRequestsPerSecond:       2,
		IPBlockDurationSeconds:    60,
		TokenRequestsPerSecond:    2,
		TokenBlockDurationSeconds: 60,
	}

	uc := NewRateLimiterUseCase(repo, &cfg)

	ctx := context.Background()

	start := time.Now()
	res, err := uc.CheckRateLimit(ctx, entities.RateLimitRequest{
		Key:       "127.0.0.1",
		Type:      entities.IPRateLimit,
		Timestamp: start,
	})

	require.NoError(t, err, "unexpected error")

	assert.True(t, res.Allowed, "expected request to be allowed")
	assert.Equal(t, 1, res.Remaining, "remaining requests should be 1")
	assert.True(t, res.ResetTime.After(start), "reset time should be after the request time")
	assert.True(t, res.BlockUntil.IsZero(), "blockUntil should be zero when request is allowed")
	assert.Equal(t, 1, repo.counts["127.0.0.1"], "repository should have recorded exactly one request")
}

func TestRateLimiter_Blocking(t *testing.T) {
	repo := newMockRepository()
	cfg := appconfig.RateLimiterConfig{
		IPRequestsPerSecond:       1,
		IPBlockDurationSeconds:    60,
		TokenRequestsPerSecond:    1,
		TokenBlockDurationSeconds: 60,
	}

	uc := NewRateLimiterUseCase(repo, &cfg)
	ctx := context.Background()

	_, _ = uc.CheckRateLimit(ctx, entities.RateLimitRequest{
		Key:       "127.0.0.1",
		Type:      entities.IPRateLimit,
		Timestamp: time.Now(),
	})

	start := time.Now()
	res, err := uc.CheckRateLimit(ctx, entities.RateLimitRequest{
		Key:       "127.0.0.1",
		Type:      entities.IPRateLimit,
		Timestamp: start,
	})
	require.NoError(t, err, "unexpected error")

	assert.False(t, res.Allowed, "expected request to be blocked on second call")
	assert.Equal(t, 0, res.Remaining, "remaining requests should be zero when blocked")

	assert.False(t, res.BlockUntil.IsZero(), "blockUntil should be set when request is blocked")
	assert.True(t, res.BlockUntil.After(start), "blockUntil should be in the future")

	blockUntil, ok := repo.blocks["127.0.0.1"]
	assert.True(t, ok, "repository should have recorded a block for the key")
	assert.Equal(t, blockUntil, res.BlockUntil, "blockUntil in response should match repository value")

}
