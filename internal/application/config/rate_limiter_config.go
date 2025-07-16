package config

type RateLimiterConfig struct {
	IPRequestsPerSecond       int
	IPBlockDurationSeconds    int
	TokenRequestsPerSecond    int
	TokenBlockDurationSeconds int
}
