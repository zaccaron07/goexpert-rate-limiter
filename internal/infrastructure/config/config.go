package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis       RedisConfig
	RateLimiter RateLimiterConfig
	Server      ServerConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RateLimiterConfig struct {
	IPRequestsPerSecond       int
	IPBlockDurationSeconds    int
	TokenRequestsPerSecond    int
	TokenBlockDurationSeconds int
}

type ServerConfig struct {
	Port string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	config := &Config{}

	config.Redis = RedisConfig{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		DB:   getEnvAsInt("REDIS_DB", 0),
	}

	config.RateLimiter = RateLimiterConfig{
		IPRequestsPerSecond:       getEnvAsInt("IP_REQUESTS_PER_SECOND", 10),
		IPBlockDurationSeconds:    getEnvAsInt("IP_BLOCK_DURATION_SECONDS", 300),
		TokenRequestsPerSecond:    getEnvAsInt("TOKEN_REQUESTS_PER_SECOND", 100),
		TokenBlockDurationSeconds: getEnvAsInt("TOKEN_BLOCK_DURATION_SECONDS", 600),
	}

	config.Server = ServerConfig{
		Port: getEnv("SERVER_PORT", "8080"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
