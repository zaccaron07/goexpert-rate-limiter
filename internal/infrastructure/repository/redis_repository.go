package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"goexpert-rate-limiter/internal/domain/entities"

	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(addr, password string, db int) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{
		client: client,
	}, nil
}

func (r *RedisRepository) GetRequestCount(ctx context.Context, key string, rateLimitType entities.RateLimitType) (int, error) {
	redisKey := r.buildKey(key, rateLimitType, "count")

	val, err := r.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to parse request count: %w", err)
	}

	return count, nil
}

func (r *RedisRepository) IncrementRequestCount(ctx context.Context, key string, rateLimitType entities.RateLimitType, expiration time.Duration) (int, error) {
	redisKey := r.buildKey(key, rateLimitType, "count")

	newCount64, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment request count: %w", err)
	}

	if newCount64 == 1 {
		if err := r.client.Expire(ctx, redisKey, expiration).Err(); err != nil {
			return 0, fmt.Errorf("failed to set expiration: %w", err)
		}
	}

	return int(newCount64), nil
}

func (r *RedisRepository) SetBlocked(ctx context.Context, key string, rateLimitType entities.RateLimitType, blockUntil time.Time) error {
	redisKey := r.buildKey(key, rateLimitType, "blocked")

	blockDuration := time.Until(blockUntil)
	if blockDuration <= 0 {
		return fmt.Errorf("invalid block duration")
	}

	err := r.client.Set(ctx, redisKey, blockUntil.Unix(), blockDuration).Err()
	if err != nil {
		return fmt.Errorf("failed to set blocked status: %w", err)
	}

	return nil
}

func (r *RedisRepository) IsBlocked(ctx context.Context, key string, rateLimitType entities.RateLimitType) (bool, time.Time, error) {
	redisKey := r.buildKey(key, rateLimitType, "blocked")

	val, err := r.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return false, time.Time{}, nil
	}
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to check blocked status: %w", err)
	}

	blockUntilUnix, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to parse block until time: %w", err)
	}

	blockUntil := time.Unix(blockUntilUnix, 0)
	return time.Now().Before(blockUntil), blockUntil, nil
}

func (r *RedisRepository) GetBlockUntil(ctx context.Context, key string, rateLimitType entities.RateLimitType) (time.Time, error) {
	redisKey := r.buildKey(key, rateLimitType, "blocked")

	val, err := r.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get block until time: %w", err)
	}

	blockUntilUnix, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse block until time: %w", err)
	}

	return time.Unix(blockUntilUnix, 0), nil
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

func (r *RedisRepository) buildKey(key string, rateLimitType entities.RateLimitType, suffix string) string {
	return fmt.Sprintf("rate_limiter:%s:%s:%s", rateLimitType, key, suffix)
}
