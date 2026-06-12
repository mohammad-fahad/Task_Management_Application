package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/taskmanager/backend/internal/models"
)

// RecoveryMiddleware is a middleware that catches panics and returns a 500 error.
type RecoveryMiddleware struct {
	// If true, the stack trace is included in the error response (dev mode only).
	IncludeStack bool
}

// NewRecoveryMiddleware creates a new RecoveryMiddleware.
func NewRecoveryMiddleware(includeStack bool) *RecoveryMiddleware {
	return &RecoveryMiddleware{IncludeStack: includeStack}
}

// Recovery returns a middleware handler that recovers from panics.
func (m *RecoveryMiddleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Log the panic
				log.Printf("PANIC recovered: %v\n%s", rec, string(debug.Stack()))

				// Build error response
				errorMsg := "internal server error"
				statusCode := http.StatusInternalServerError

				// Check for specific panic types
				switch v := rec.(type) {
				case *httpError:
					statusCode = v.StatusCode
					errorMsg = v.Message
				case error:
					errorMsg = v.Error()
				case string:
					errorMsg = v
				}

				resp := models.APIResponse{
					Success: false,
					Error:   errorMsg,
				}

				if m.IncludeStack {
					resp.Data = map[string]string{
						"stack": string(debug.Stack()),
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)
				json.NewEncoder(w).Encode(resp)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// httpError is a typed panic value for HTTP errors.
type httpError struct {
	StatusCode int
	Message    string
}

// PanicForError causes a panic that will be recovered as an HTTP error response.
// This is useful in handlers for aborting early with a specific status code.
func PanicForError(statusCode int, message string) {
	panic(&httpError{StatusCode: statusCode, Message: message})
}

// RequestLoggingMiddleware logs incoming HTTP requests.
func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("→ %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}