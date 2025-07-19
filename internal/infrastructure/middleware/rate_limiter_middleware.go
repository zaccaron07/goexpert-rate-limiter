package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goexpert-rate-limiter/internal/application/usecases"
	"goexpert-rate-limiter/internal/domain/entities"
)

func RateLimiterMiddleware(rateLimiterUC usecases.RateLimiterUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			key, limitType, err := extractRateLimitKey(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			result, err := rateLimiterUC.CheckRateLimit(ctx, entities.RateLimitRequest{
				Key:       key,
				Type:      limitType,
				Timestamp: time.Now(),
			})

			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !result.Allowed {
				handleRateLimitExceeded(w, result)
				return
			}

			addRateLimitHeaders(w, result)
			next.ServeHTTP(w, r)
		})
	}
}

func handleRateLimitExceeded(w http.ResponseWriter, result entities.RateLimitResult) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"error":       "you have reached the maximum number of requests or actions allowed within a certain time frame",
		"block_until": result.BlockUntil.Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

func addRateLimitHeaders(w http.ResponseWriter, result entities.RateLimitResult) {
	w.Header().Set("X-Ratelimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("X-Ratelimit-Reset", result.ResetTime.Format(time.RFC3339))
}

func getTokenFromRequest(r *http.Request) (string, bool) {
	apiKey := r.Header.Get("API_KEY")
	if apiKey == "" {
		return "", false
	}
	return apiKey, true
}

func getIPFromRequest(r *http.Request) (string, error) {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0]), nil
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP, nil
	}

	remoteAddr := r.RemoteAddr
	if remoteAddr == "" {
		return "", fmt.Errorf("could not determine IP address")
	}

	if colonIndex := strings.LastIndex(remoteAddr, ":"); colonIndex != -1 {
		remoteAddr = remoteAddr[:colonIndex]
	}

	return remoteAddr, nil
}

func extractRateLimitKey(r *http.Request) (string, entities.RateLimitType, error) {
	if token, ok := getTokenFromRequest(r); ok {
		return token, entities.TokenRateLimit, nil
	}

	ip, err := getIPFromRequest(r)
	if err != nil {
		return "", "", err
	}
	return ip, entities.IPRateLimit, nil
}
