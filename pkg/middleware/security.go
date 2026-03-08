package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// SecurityHeaders middleware adds security headers to all responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; img-src 'self' data: https:; font-src 'self' data: https://cdnjs.cloudflare.com; connect-src 'self';")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS (only in production)
		// if config.AppConfig.IsProduction() {
		// 	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// }

		// Remove server information
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")

		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements basic rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimitMiddleware applies rate limiting to requests
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := getClientIP(r)

		// Check rate limit
		if !rl.isAllowed(clientIP) {
			// Check if it's an API request or HTML request
			acceptHeader := r.Header.Get("Accept")
			if strings.Contains(acceptHeader, "application/json") || strings.Contains(r.URL.Path, "/api/") {
				// API request - return JSON
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Rate limit exceeded. Please try again later."}`))
			} else {
				// HTML request - redirect with error message based on the route
				var redirectPath string
				if strings.Contains(r.URL.Path, "/forget-password") || strings.Contains(r.URL.Path, "/reset-password") {
					redirectPath = "/forget-password?error=rate_limit_exceeded"
				} else {
					redirectPath = "/login?error=rate_limit_exceeded"
				}
				http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowed checks if the request is allowed based on rate limiting rules
func (rl *RateLimiter) isAllowed(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this IP
	requests, exists := rl.requests[clientIP]
	if !exists {
		requests = []time.Time{}
	}

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[clientIP] = validRequests

	return true
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIndex := strings.Index(ip, ","); commaIndex != -1 {
			ip = strings.TrimSpace(ip[:commaIndex])
		}
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback to remote address
	// Remove port if present
	if colonIndex := strings.LastIndex(r.RemoteAddr, ":"); colonIndex != -1 {
		return r.RemoteAddr[:colonIndex]
	}
	return r.RemoteAddr
}

// CleanupRateLimiter periodically cleans up old rate limiting data
func (rl *RateLimiter) CleanupRateLimiter() {
	ticker := time.NewTicker(rl.window)
	go func() {
		for range ticker.C {
			rl.cleanup()
		}
	}()
}

// cleanup removes old rate limiting data
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}
