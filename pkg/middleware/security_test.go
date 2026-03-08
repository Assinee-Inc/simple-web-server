package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	// Apply security headers middleware
	middleware := SecurityHeaders(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Execute request
	middleware.ServeHTTP(w, req)

	// Check security headers
	headers := w.Header()

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range expectedHeaders {
		if value := headers.Get(header); value != expectedValue {
			t.Errorf("Expected header %s to be %s, got %s", header, expectedValue, value)
		}
	}

	// Check that server information headers are removed
	if headers.Get("Server") != "" {
		t.Error("Server header should be removed")
	}
	if headers.Get("X-Powered-By") != "" {
		t.Error("X-Powered-By header should be removed")
	}
}

func TestRateLimiter(t *testing.T) {
	// Create rate limiter with limit of 2 requests
	limiter := NewRateLimiter(2, time.Minute)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := limiter.RateLimitMiddleware(handler)

	// Create request
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	// First request should succeed
	w1 := httptest.NewRecorder()
	middleware.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w1.Code)
	}

	// Second request should succeed
	w2 := httptest.NewRecorder()
	middleware.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w2.Code)
	}

	// Third request should be rate limited (redirect for HTML requests)
	w3 := httptest.NewRecorder()
	middleware.ServeHTTP(w3, req)
	if w3.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 (redirect), got %d", w3.Code)
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	// Create rate limiter with short window
	limiter := NewRateLimiter(1, 100*time.Millisecond)

	// Make a request
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := limiter.RateLimitMiddleware(handler)

	middleware.ServeHTTP(w, req)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Manually trigger cleanup
	limiter.cleanup()

	// Check that data was cleaned up
	limiter.mutex.RLock()
	_, exists := limiter.requests["127.0.0.1:12345"]
	limiter.mutex.RUnlock()

	if exists {
		t.Error("Rate limiter data should be cleaned up")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			remoteAddr: "127.0.0.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			remoteAddr: "127.0.0.1:12345",
			expected:   "10.0.0.1",
		},
		{
			name:       "No headers, use remote addr (port removed)",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:12345",
			expected:   "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := getClientIP(req)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
