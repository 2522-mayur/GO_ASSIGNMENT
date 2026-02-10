package services

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"taskapi/config"
	"taskapi/database"
	"taskapi/middleware"
	"taskapi/models"
	"taskapi/repositories"
)

// UserService handles user-related business logic
type UserService struct {
	db  *database.DB
	cfg *config.Config
}

// NewUserService creates a new user service
func NewUserService(db *database.DB, cfg *config.Config) *UserService {
	return &UserService{db: db, cfg: cfg}
}

// Register creates a new user
func (s *UserService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return nil, errors.New("email, username, and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := repositories.CreateUser(s.db, user); err != nil {
		return nil, errors.New("user already exists or database error")
	}

	token, err := middleware.GenerateToken(user, s.cfg)
	if err != nil {
		return nil, err
	}

	// Don't expose password in response
	user.Password = ""

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login authenticates a user
func (s *UserService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := repositories.GetUserByEmail(s.db, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := middleware.GenerateToken(user, s.cfg)
	if err != nil {
		return nil, err
	}

	// Don't expose password in response
	user.Password = ""

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// TaskService handles task-related business logic
type TaskService struct {
	db *database.DB
}

// NewTaskService creates a new task service
func NewTaskService(db *database.DB) *TaskService {
	return &TaskService{db: db}
}

// CreateTask creates a new task for a user
func (s *TaskService) CreateTask(userID string, req *models.CreateTaskRequest) (*models.Task, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	task := &models.Task{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
	}

	if err := repositories.CreateTask(s.db, task); err != nil {
		return nil, err
	}

	// Don't expose UserID in response
	task.UserID = ""
	return task, nil
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(taskID string) (*models.Task, error) {
	task, err := repositories.GetTaskByID(s.db, taskID)
	if err != nil {
		return nil, err
	}
	task.UserID = ""
	return task, nil
}

// GetUserTasks retrieves all tasks for a user
func (s *TaskService) GetUserTasks(userID string) ([]*models.Task, error) {
	tasks, err := repositories.GetUserTasks(s.db, userID)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		task.UserID = ""
	}
	return tasks, nil
}

// GetAllTasks retrieves all tasks (for admin)
func (s *TaskService) GetAllTasks() ([]*models.Task, error) {
	tasks, err := repositories.GetAllTasks(s.db)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		task.UserID = ""
	}
	return tasks, nil
}

// UpdateTask updates a task
func (s *TaskService) UpdateTask(userID string, taskID string, req *models.UpdateTaskRequest, isAdmin bool) (*models.Task, error) {
	task, err := repositories.GetTaskByID(s.db, taskID)
	if err != nil {
		return nil, err
	}

	// Check authorization (user can only update their own tasks, unless admin)
	if !isAdmin && task.UserID != userID {
		return nil, errors.New("unauthorized to update this task")
	}

	// Validate status
	validStatuses := map[string]bool{"pending": true, "in_progress": true, "completed": true}
	if req.Status != "" && !validStatuses[req.Status] {
		return nil, errors.New("invalid status")
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}

	if err := repositories.UpdateTask(s.db, task); err != nil {
		return nil, err
	}

	task.UserID = ""
	return task, nil
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(userID string, taskID string, isAdmin bool) error {
	task, err := repositories.GetTaskByID(s.db, taskID)
	if err != nil {
		return err
	}

	// Check authorization
	if !isAdmin && task.UserID != userID {
		return errors.New("unauthorized to delete this task")
	}

	return repositories.DeleteTask(s.db, taskID)
}
