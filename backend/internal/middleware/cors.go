package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// defaultAllowedOrigins are always included regardless of environment configuration.
var defaultAllowedOrigins = []string{
	"http://localhost:3000",
	"https://task-management-application-omega-lyart.vercel.app",
}

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers.
//
// It reads the API_ALLOWED_ORIGINS environment variable for a comma-separated
// list of additional allowed origins, then always appends the built-in defaults
// (local dev + production Vercel domain) to form the complete allow-list.
//
// Preflight OPTIONS requests are terminated immediately with HTTP 200 OK and
// never forwarded to route handlers, authentication checks, or the business
// logic layer. This eliminates unnecessary latency and ensures that credentialled
// cross-origin requests (cookies) are not blocked by auth middleware during the
// preflight phase.
//
// CRITICAL: Wildcard origins (*) are incompatible with
// Access-Control-Allow-Credentials: true per the Fetch Standard. If a wildcard
// is detected in the environment variable, it is silently ignored and the
// default origins are used instead.
func CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := buildAllowedOrigins()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Determine if the request's origin is permitted
		allowedOrigin := resolveAllowedOrigin(origin, allowedOrigins)

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Terminate preflight requests immediately — do NOT pass to auth middleware
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// buildAllowedOrigins merges environment-configured origins with the built-in defaults.
// It never returns a wildcard (*).
func buildAllowedOrigins() []string {
	// Start with the built-in defaults
	originSet := make(map[string]struct{})
	for _, o := range defaultAllowedOrigins {
		originSet[o] = struct{}{}
	}

	// Merge in origins from the environment variable
	envOrigins := os.Getenv("API_ALLOWED_ORIGINS")
	if envOrigins != "" {
		for _, o := range strings.Split(envOrigins, ",") {
			o = strings.TrimSpace(o)
			if o != "" && o != "*" {
				originSet[o] = struct{}{}
			}
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(originSet))
	for o := range originSet {
		result = append(result, o)
	}

	log.Printf("[CORS] Allowed origins: %v", result)
	return result
}

// resolveAllowedOrigin checks if the request origin is in the allowed list.
// Returns the origin if allowed, or an empty string if denied.
func resolveAllowedOrigin(requestOrigin string, allowedOrigins []string) string {
	if requestOrigin == "" {
		return ""
	}

	for _, allowed := range allowedOrigins {
		if strings.EqualFold(requestOrigin, allowed) {
			return allowed
		}
	}

	return ""
}