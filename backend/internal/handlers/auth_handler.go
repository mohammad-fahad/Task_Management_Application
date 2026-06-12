package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/taskmanager/backend/internal/middleware"
	"github.com/taskmanager/backend/internal/models"
	"github.com/taskmanager/backend/internal/service"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeValidationError(w, "invalid request body", models.ErrorDetail{
			Field:   "body",
			Message: "failed to parse JSON request body",
		})
		return
	}

	if err := h.validate.Struct(&input); err != nil {
		writeValidationErrors(w, err)
		return
	}

	user, err := h.authService.Register(r.Context(), &input)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyInUse) {
			writeValidationError(w, "registration failed", models.ErrorDetail{
				Field:   "email",
				Message: "email address is already registered",
			})
			return
		}
		writeInternalError(w, "registration failed")
		return
	}

	h.authService.SetAuthCookie(w, "")
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "user registered successfully",
		Data:    user,
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeValidationError(w, "invalid request body", models.ErrorDetail{
			Field:   "body",
			Message: "failed to parse JSON request body",
		})
		return
	}

	if err := h.validate.Struct(&input); err != nil {
		writeValidationErrors(w, err)
		return
	}

	user, token, err := h.authService.Login(r.Context(), &input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeUnauthorizedError(w, "invalid email or password")
			return
		}
		writeInternalError(w, "login failed")
		return
	}

	h.authService.SetAuthCookie(w, token)
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "login successful",
		Data: map[string]interface{}{
			"user":  user,
			"token": token,
		},
	})
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authService.ClearAuthCookie(w)
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "logged out successfully",
	})
}

// Me handles GET /api/v1/auth/me – returns the current authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeUnauthorizedError(w, "not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeInternalError(w, "failed to fetch user")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}
