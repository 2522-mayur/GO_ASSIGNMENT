package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"taskapi/middleware"
	"taskapi/models"
	"taskapi/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.userService.Register(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.userService.Login(&req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// TaskHandler handles task endpoints
type TaskHandler struct {
	taskService *services.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *services.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// CreateTask handles task creation
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := h.taskService.CreateTask(claims.UserID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

// GetTask handles getting a single task
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	// Check authorization
	if claims.Role != "admin" && task.UserID != "" && task.UserID != claims.UserID {
		writeError(w, http.StatusForbidden, "Unauthorized to access this task")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// GetTasks handles getting all tasks for the user or all tasks if admin
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var tasks []*models.Task
	var err error

	if claims.Role == "admin" {
		tasks, err = h.taskService.GetAllTasks()
	} else {
		tasks, err = h.taskService.GetUserTasks(claims.UserID)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error retrieving tasks")
		return
	}

	if tasks == nil {
		tasks = []*models.Task{}
	}

	writeJSON(w, http.StatusOK, tasks)
}

// UpdateTask handles task updates
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := h.taskService.UpdateTask(claims.UserID, taskID, &req, claims.Role == "admin")
	if err != nil {
		if err.Error() == "unauthorized to update this task" {
			writeError(w, http.StatusForbidden, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// DeleteTask handles task deletion
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]

	err := h.taskService.DeleteTask(claims.UserID, taskID, claims.Role == "admin")
	if err != nil {
		if err.Error() == "unauthorized to delete this task" {
			writeError(w, http.StatusForbidden, err.Error())
		} else {
			writeError(w, http.StatusNotFound, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

// Helper functions

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{Error: message})
}
