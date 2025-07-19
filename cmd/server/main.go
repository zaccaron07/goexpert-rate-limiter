package main

import (
	"fmt"
	"log"
	"net/http"

	appconfig "goexpert-rate-limiter/internal/application/config"
	usecases "goexpert-rate-limiter/internal/application/usecases"
	"goexpert-rate-limiter/internal/infrastructure/config"
	"goexpert-rate-limiter/internal/infrastructure/middleware"
	redisrepo "goexpert-rate-limiter/internal/infrastructure/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configurations: %v", err)
	}

	repo, err := redisrepo.NewRedisRepository(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	defer repo.Close()

	rlCfg := appconfig.RateLimiterConfig{
		IPRequestsPerSecond:       cfg.RateLimiter.IPRequestsPerSecond,
		IPBlockDurationSeconds:    cfg.RateLimiter.IPBlockDurationSeconds,
		TokenRequestsPerSecond:    cfg.RateLimiter.TokenRequestsPerSecond,
		TokenBlockDurationSeconds: cfg.RateLimiter.TokenBlockDurationSeconds,
	}

	ratelimiterUC := usecases.NewRateLimiterUseCase(repo, &rlCfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello, world! You are not rate limited."}`))
	})

	rateLimitedHandler := middleware.RateLimiterMiddleware(ratelimiterUC)(mux)

	addr := ":" + cfg.Server.Port
	fmt.Printf("Server started at port %s\n", cfg.Server.Port)
	fmt.Printf("IP Requests Per Second: %d\n", rlCfg.IPRequestsPerSecond)
	fmt.Printf("IP Block Duration Seconds: %d\n", rlCfg.IPBlockDurationSeconds)
	fmt.Printf("Token Requests Per Second: %d\n", rlCfg.TokenRequestsPerSecond)
	fmt.Printf("Token Block Duration Seconds: %d\n", rlCfg.TokenBlockDurationSeconds)
	if err := http.ListenAndServe(addr, rateLimitedHandler); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
