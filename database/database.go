package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"taskapi/config"
)

// DB holds the database connection
type DB struct {
	Conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{Conn: conn}, nil
}

// RunMigrations creates the necessary database tables
func (db *DB) RunMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user',
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);`,
	}

	for _, migration := range migrations {
		if _, err := db.Conn.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.Conn.Close()
}
