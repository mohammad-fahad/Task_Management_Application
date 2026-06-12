package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/taskmanager/backend/internal/models"
	"github.com/taskmanager/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrEmailAlreadyInUse  = errors.New("email already registered")
	ErrUserNotAuthorized  = errors.New("user not authorized")
)

// AuthService handles authentication and authorization business logic.
type AuthService struct {
	userRepo   repository.UserRepository
	jwtSecret  []byte
	cookieName string
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		cookieName: "auth_token",
	}
}

// Register creates a new user account with hashed password.
func (s *AuthService) Register(ctx context.Context, input *models.UserInput) (*models.User, error) {
	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyInUse
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	input.Password = string(hashedPassword)

	user, err := s.userRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Clear password hash from returned user (json:"-" handles this, but be explicit)
	user.PasswordHash = ""
	return user, nil
}

// Login authenticates a user and returns JWT tokens.
func (s *AuthService) Login(ctx context.Context, input *models.LoginInput) (*models.User, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	user.PasswordHash = ""
	return user, token, nil
}

// ValidateToken validates a JWT token and returns the user ID and role.
func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, "", ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}

	role, _ := claims["role"].(string)

	return userID, role, nil
}

// SetAuthCookie sets the JWT as an HTTP-Only cookie on the response.
func (s *AuthService) SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(24 * time.Hour.Seconds()), // 24 hours
	})
}

// ClearAuthCookie removes the auth cookie (for logout).
func (s *AuthService) ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
}

// GetCookieName returns the name of the auth cookie.
func (s *AuthService) GetCookieName() string {
	return s.cookieName
}

// GetUserByID retrieves a user by their ID (used by the /me endpoint).
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	user.PasswordHash = ""
	return user, nil
}

// generateToken creates a new JWT token for the given user.
func (s *AuthService) generateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
		"iss":   "taskmanager-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}