package entities

import (
	"time"
)

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	ResetTime  time.Time
	BlockUntil time.Time
}

type RateLimitType string

const (
	IPRateLimit    RateLimitType = "ip"
	TokenRateLimit RateLimitType = "token"
)

type RateLimitRequest struct {
	Key       string
	Type      RateLimitType
	Timestamp time.Time
}

type RateLimitResponse struct {
	Allowed    bool      `json:"allowed"`
	Remaining  int       `json:"remaining"`
	ResetTime  time.Time `json:"reset_time"`
	BlockUntil time.Time `json:"block_until,omitempty"`
}
