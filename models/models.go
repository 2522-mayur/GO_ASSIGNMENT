package models

import "time"

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"-"` // Never expose password in JSON
	Role     string `json:"role"` // "user" or "admin"
	CreatedAt time.Time `json:"created_at"`
}

// Task represents a task
type Task struct {
	ID        string `json:"id"`
	UserID    string `json:"-"` // Don't expose in JSON
	Title     string `json:"title"`
	Description string `json:"description"`
	Status    string `json:"status"` // pending, in_progress, completed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTaskRequest is the request body for creating a task
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTaskRequest is the request body for updating a task
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the response for authentication
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
