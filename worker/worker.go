package worker

import (
	"log"
	"sync"
	"time"

	"taskapi/config"
	"taskapi/database"
	"taskapi/repositories"
)

// TaskWorker handles background task auto-completion
type TaskWorker struct {
	db              *database.DB
	cfg             *config.Config
	taskChannel     chan string
	stopChannel     chan struct{}
	wg              sync.WaitGroup
	mu              sync.Mutex
	processedTasks  map[string]bool
}

// NewTaskWorker creates a new task worker
func NewTaskWorker(db *database.DB, cfg *config.Config) *TaskWorker {
	return &TaskWorker{
		db:             db,
		cfg:            cfg,
		taskChannel:    make(chan string, 100), // buffered channel
		stopChannel:    make(chan struct{}),
		processedTasks: make(map[string]bool),
	}
}

// Start starts the background worker
func (w *TaskWorker) Start() {
	log.Println("Starting task auto-completion worker...")

	// Start worker goroutine to process tasks from channel
	w.wg.Add(1)
	go w.processTasksFromChannel()

	// Start checker goroutine to periodically find and send tasks for auto-completion
	w.wg.Add(1)
	go w.checkAndQueueTasks()

	log.Println("Task worker started successfully")
}

// Stop stops the background worker gracefully
func (w *TaskWorker) Stop() {
	log.Println("Stopping task worker...")
	close(w.stopChannel)
	w.wg.Wait()
	close(w.taskChannel)
	log.Println("Task worker stopped")
}

// checkAndQueueTasks periodically checks for tasks that should be auto-completed
func (w *TaskWorker) checkAndQueueTasks() {
	defer w.wg.Done()

	// Check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChannel:
			return
		case <-ticker.C:
			w.findAndQueueTasks()
		}
	}
}

// findAndQueueTasks finds tasks that need auto-completion and sends them to the channel
func (w *TaskWorker) findAndQueueTasks() {
	tasks, err := repositories.GetTasksForAutoCompletion(w.db, w.cfg.AutoCompleteMinutes)
	if err != nil {
		log.Printf("Error fetching tasks for auto-completion: %v\n", err)
		return
	}

	for _, task := range tasks {
		// Only process each task once
		w.mu.Lock()
		if _, exists := w.processedTasks[task.ID]; !exists {
			w.processedTasks[task.ID] = true
			w.mu.Unlock()

			// Send task ID to channel (non-blocking with timeout)
			select {
			case w.taskChannel <- task.ID:
				log.Printf("Queued task %s for auto-completion\n", task.ID)
			case <-time.After(100 * time.Millisecond):
				// Channel full, try again next time
				w.mu.Lock()
				delete(w.processedTasks, task.ID)
				w.mu.Unlock()
			}
		} else {
			w.mu.Unlock()
		}
	}
}

// processTasksFromChannel processes tasks from the channel
func (w *TaskWorker) processTasksFromChannel() {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopChannel:
			return
		case taskID := <-w.taskChannel:
			w.autoCompleteTask(taskID)
		}
	}
}

// autoCompleteTask marks a task as completed
func (w *TaskWorker) autoCompleteTask(taskID string) {
	// Verify the task still exists and is not already completed
	task, err := repositories.GetTaskByID(w.db, taskID)
	if err != nil {
		log.Printf("Task %s not found: %v\n", taskID, err)
		return
	}

	// Double-check status (in case it was manually completed)
	if task.Status == "completed" {
		log.Printf("Task %s is already completed, skipping auto-completion\n", taskID)
		return
	}

	// Auto-complete the task
	if err := repositories.AutoCompleteTask(w.db, taskID); err != nil {
		log.Printf("Error auto-completing task %s: %v\n", taskID, err)
		return
	}

	log.Printf("Task %s auto-completed successfully\n", taskID)
}

// SubmitTask allows external submission of tasks to be processed
func (w *TaskWorker) SubmitTask(taskID string) error {
	select {
	case w.taskChannel <- taskID:
		log.Printf("Manually submitted task %s for processing\n", taskID)
		return nil
	case <-time.After(5 * time.Second):
		return ErrChannelFull
	}
}

// ErrChannelFull is returned when the task channel is full
var ErrChannelFull = &ChannelFullError{}

type ChannelFullError struct{}

func (e *ChannelFullError) Error() string {
	return "task queue is full"
}
