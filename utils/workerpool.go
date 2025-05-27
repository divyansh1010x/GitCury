package utils

import (
	"context"
	"sync"
	"time"
)

// WorkerPool implements a reusable worker pool with configurable concurrency limit
type WorkerPool struct {
	maxWorkers int
	semaphore  chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	errors     []error
}

// NewWorkerPool creates a new worker pool with the specified maximum concurrent workers
func NewWorkerPool(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	return &WorkerPool{
		maxWorkers: maxWorkers,
		semaphore:  make(chan struct{}, maxWorkers),
		errors:     make([]error, 0),
	}
}

// Submit adds a task to the worker pool with the specified timeout
func (wp *WorkerPool) Submit(taskName string, timeout time.Duration, task func() error) {
	wp.wg.Add(1)

	go func() {
		defer wp.wg.Done()

		// Acquire semaphore slot (blocks if max workers is reached)
		wp.semaphore <- struct{}{}
		defer func() { <-wp.semaphore }()

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Channel for task completion
		done := make(chan error, 1)

		// Run the task in a separate goroutine
		go func() {
			err := SafeExecute(taskName, task)
			done <- err
		}()

		// Wait for either task completion or timeout
		select {
		case err := <-done:
			if err != nil {
				wp.addError(err)
			}
		case <-ctx.Done():
			wp.addError(NewSystemError(
				"Task timed out",
				ctx.Err(),
				map[string]interface{}{
					"taskName": taskName,
					"timeout":  timeout.String(),
				},
			))
		}
	}()
}

// Wait blocks until all tasks have completed
func (wp *WorkerPool) Wait() []error {
	wp.wg.Wait()
	return wp.Errors()
}

// Errors returns all errors collected from task execution
func (wp *WorkerPool) Errors() []error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Make a copy to avoid race conditions
	errCopy := make([]error, len(wp.errors))
	copy(errCopy, wp.errors)

	return errCopy
}

// HasErrors returns true if any tasks have reported errors
func (wp *WorkerPool) HasErrors() bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return len(wp.errors) > 0
}

// addError safely adds an error to the error list
func (wp *WorkerPool) addError(err error) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.errors = append(wp.errors, err)
}
