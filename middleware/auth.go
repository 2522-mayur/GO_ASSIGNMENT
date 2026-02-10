package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"taskapi/config"
	"taskapi/models"
)

const (
	AuthContextKey = "user"
	BearerScheme   = "Bearer"
)

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token
func GenerateToken(user *models.User, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpiryHours) * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(tokenString string, cfg *config.Config) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// AuthMiddleware is a middleware that checks for valid JWT token
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != BearerScheme {
				writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			claims, err := ValidateToken(parts[1], cfg)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user claims from context
func GetUserFromContext(r *http.Request) *Claims {
	claims := r.Context().Value(AuthContextKey)
	if claims == nil {
		return nil
	}
	return claims.(*Claims)
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeError writes an error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
