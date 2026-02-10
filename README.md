# Task Management API

A simple Go REST API for task management with JWT authentication, PostgreSQL database, and automatic task completion using goroutines and channels.

## Features

- **JWT Authentication**: Secure token-based authentication with configurable expiry
- **Authorization**: Users can access only their own tasks; admins can access all
- **Task Management**: Create, read, update, and delete tasks
- **Background Worker**: Automatic task completion after X minutes using goroutines
- **PostgreSQL**: Persistent data storage with proper database design
- **Error Handling**: Proper HTTP status codes and JSON error responses
- **Thread-Safe**: Concurrent task processing with channels and mutex protection

## Project Structure

```
.
├── config/          # Configuration management
├── database/        # Database connection and migrations
├── handlers/        # HTTP request handlers
├── middleware/      # JWT authentication middleware
├── models/          # Data models
├── repositories/    # Database access layer
├── services/        # Business logic layer
├── worker/          # Background task worker
├── main.go         # Application entry point
├── docker-compose.yml
├── .env.example
└── go.mod
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher (or Docker)
- Git

## Installation & Setup

### 1. Clone the Repository

```bash
cd /workspaces/GO_ASSIGNMENT
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up PostgreSQL

**Option A: Using Docker Compose (Recommended)**

```bash
docker-compose up -d
```

**Option B: Local PostgreSQL**

Make sure PostgreSQL is running and create a database:

```bash
psql -U postgres -c "CREATE DATABASE taskdb;"
```

### 4. Configure Environment Variables

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

Edit `.env` with your configuration (defaults work fine for Docker setup):

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=taskdb
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRY_HOURS=24
AUTO_COMPLETE_MINUTES=30
SERVER_PORT=8080
```

### 5. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication

#### Register User

```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "username",
  "password": "password123"
}
```

Response:
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Login User

```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

### Tasks (Protected - Requires JWT Token)

Add `Authorization: Bearer <token>` header to all requests.

#### Create Task

```bash
POST /api/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "My Task",
  "description": "Task description"
}
```

#### Get All Tasks

```bash
GET /api/tasks
Authorization: Bearer <token>
```

- Regular users get only their own tasks
- Admin users get all tasks

#### Get Single Task

```bash
GET /api/tasks/{id}
Authorization: Bearer <token>
```

#### Update Task

```bash
PUT /api/tasks/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "status": "in_progress"
}
```

Valid statuses: `pending`, `in_progress`, `completed`

#### Delete Task

```bash
DELETE /api/tasks/{id}
Authorization: Bearer <token>
```

#### Health Check

```bash
GET /health
```

## How It Works

### Authentication Flow

1. User registers or logs in
2. Server validates credentials and generates JWT token
3. Token expires after `JWT_EXPIRY_HOURS` (default: 24 hours)
4. All protected endpoints require valid token in `Authorization: Bearer <token>` header

### Background Task Worker

The task worker runs continuously in the background:

1. **Checker Goroutine**: Runs every minute to find tasks older than `AUTO_COMPLETE_MINUTES`
2. **Queue Channel**: Found tasks are sent to a buffered channel (capacity: 100)
3. **Processor Goroutine**: Processes tasks from the channel
4. **Thread Safety**: Uses mutex to track processed tasks and prevent duplicates
5. **Database Update**: Marks eligible tasks as `completed` with updated timestamp

**Auto-completion Rules:**
- Only processes tasks with status `pending` or `in_progress`
- Skips if task is already `completed`
- Skips if task was deleted
- Configurable delay via `AUTO_COMPLETE_MINUTES` environment variable

### Error Handling

All error responses follow this format:

```json
{
  "error": "Error description"
}
```

HTTP Status Codes:
- `200 OK`: Successful request
- `201 Created`: Resource created
- `400 Bad Request`: Invalid input or validation error
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: User not authorized to access resource
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Example Usage

### Complete User Flow

```bash
# 1. Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "username": "alice",
    "password": "securepass123"
  }'

# Response: {"token": "eyJhbGc...", "user": {...}}

# 2. Create a task (use token from response)
curl -X POST http://localhost:8080/api/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project",
    "description": "Finish the Go API project"
  }'

# 3. Get all tasks
curl -X GET http://localhost:8080/api/tasks \
  -H "Authorization: Bearer <token>"

# 4. Update task status
curl -X PUT http://localhost:8080/api/tasks/{task-id} \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress"
  }'

# 5. The task will auto-complete after 30 minutes (if AUTO_COMPLETE_MINUTES=30)
```

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | Database host |
| DB_PORT | 5432 | Database port |
| DB_USER | postgres | Database user |
| DB_PASSWORD | postgres | Database password |
| DB_NAME | taskdb | Database name |
| JWT_SECRET | secret-key | Secret key for JWT signing (change in production!) |
| JWT_EXPIRY_HOURS | 24 | JWT token expiry in hours |
| AUTO_COMPLETE_MINUTES | 30 | Minutes before pending tasks auto-complete |
| SERVER_PORT | 8080 | Server port |

## Development

### Running Tests

```bash
go test ./...
```

### Code Structure

- **Models**: Define data structures and request/response types
- **Database**: Handle database connections and migrations
- **Repositories**: Data access layer using SQL queries
- **Services**: Business logic and validation
- **Handlers**: HTTP request/response handling
- **Middleware**: JWT authentication and authorization
- **Worker**: Background processing with goroutines

### Key Design Decisions

1. **Separation of Concerns**: Clear boundaries between data, business logic, and HTTP handling
2. **Thread-Safe Worker**: Uses channels for safe concurrent task processing
3. **Idiomatic Go**: Follows Go conventions and best practices
4. **Error Handling**: Explicit error handling without panics
5. **Stateless API**: Each request is independent except for user context
6. **Database Indexes**: Indexes on user_id and status for query performance

## Troubleshooting

### Database Connection Error

Make sure PostgreSQL is running:
```bash
docker-compose ps
```

### JWT Token Expired

Get a new token by logging in again.

### Tasks Not Auto-Completing

Check:
1. Worker is running (logs show "Starting task auto-completion worker")
2. Tasks have status `pending` or `in_progress`
3. Tasks are older than `AUTO_COMPLETE_MINUTES`
4. No errors in logs

## Stopping the Server

Press `Ctrl+C` to gracefully shut down. The worker will:
1. Stop accepting new tasks
2. Complete processing of queued tasks
3. Exit cleanly

## License

MIT

## Support

For issues or questions, please create an issue in the repository.
