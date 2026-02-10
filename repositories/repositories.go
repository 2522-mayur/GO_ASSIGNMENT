package repositories

import (
	"database/sql"
	"errors"
	"taskapi/database"
	"taskapi/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func CreateUser(db *database.DB, user *models.User) error {
	query := `
		INSERT INTO users (email, username, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	row := db.Conn.QueryRow(query, user.Email, user.Username, user.Password, user.Role)
	return row.Scan(&user.ID, &user.CreatedAt)
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(db *database.DB, email string) (*models.User, error) {
	query := `SELECT id, email, username, password, role, created_at FROM users WHERE email = $1`

	user := &models.User{}
	row := db.Conn.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	return user, err
}

// GetUserByID retrieves a user by ID (package-level helper)
func GetUserByID(db *database.DB, id string) (*models.User, error) {
	query := `SELECT id, email, username, password, role, created_at FROM users WHERE id = $1`

	user := &models.User{}
	row := db.Conn.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	return user, err
}

// TaskRepository handles task database operations
type TaskRepository struct {
	db *database.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *database.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// CreateTask creates a new task
func CreateTask(db *database.DB, task *models.Task) error {
	query := `
		INSERT INTO tasks (user_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	row := db.Conn.QueryRow(query, task.UserID, task.Title, task.Description, "pending")
	return row.Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

// GetTaskByID retrieves a task by ID
func GetTaskByID(db *database.DB, taskID string) (*models.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks WHERE id = $1
	`

	task := &models.Task{}
	row := db.Conn.QueryRow(query, taskID)
	err := row.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("task not found")
	}

	return task, err
}

// GetUserTasks retrieves all tasks for a user
func GetUserTasks(db *database.DB, userID string) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetAllTasks retrieves all tasks (for admin)
func GetAllTasks(db *database.DB) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks ORDER BY created_at DESC
	`

	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates a task
func UpdateTask(db *database.DB, task *models.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	row := db.Conn.QueryRow(query, task.Title, task.Description, task.Status, task.ID)
	return row.Scan(&task.UpdatedAt)
}

// DeleteTask deletes a task
func DeleteTask(db *database.DB, taskID string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := db.Conn.Exec(query, taskID)
	return err
}

// GetTasksForAutoCompletion retrieves tasks that need auto-completion
func GetTasksForAutoCompletion(db *database.DB, minutes int) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE status IN ('pending', 'in_progress')
		AND created_at < NOW() - INTERVAL '1 minute' * $1
	`

	rows, err := db.Conn.Query(query, minutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// AutoCompleteTask marks a task as completed
func AutoCompleteTask(db *database.DB, taskID string) error {
	query := `
		UPDATE tasks
		SET status = 'completed', updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'in_progress')
	`
	_, err := db.Conn.Exec(query, taskID)
	return err
}
