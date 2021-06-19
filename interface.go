package tommyqueue

import (
	"context"
	"sync"
	"time"
)

type Logger interface {
	// Debug logs a message at Debug level.
	Debug(args ...interface{})

	// Info logs a message at Info level.
	Info(args ...interface{})

	Infof(format string, args ...interface{})

	// Warn logs a message at Warning level.
	Warn(args ...interface{})

	Warnf(format string, args ...interface{})

	// Error logs a message at Error level.
	Error(args ...interface{})

	Errorf(format string, args ...interface{})
}

type Worker interface {
	Start(ctx context.Context, wg *sync.WaitGroup)
	Stop(ctx context.Context)
}

// A ProcessHandler processes tasks.
//
// ProcessTask should return nil if the processing of a task
// is successful.
//
// If ProcessTask return a non-nil error or panics, the task
// will be retried after delay.
// One exception to this rule is when ProcessTask returns SkipRetry error.
// If the returned error is SkipRetry or the error wraps SkipRetry, retry is
// skipped and task will be archived instead.
type ProcessHandler interface {
	ProcessTask(context.Context, *Task) error
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as a Handler. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(context.Context, *Task) error

// ProcessTask calls fn(ctx, task)
func (fn HandlerFunc) ProcessTask(ctx context.Context, task *Task) error {
	return fn(ctx, task)
}

// An ErrorHandler handles an error occured during task processing.
type ErrorHandler interface {
	HandleError(ctx context.Context, task *Task, err error)
}

// The ErrorHandlerFunc type is an adapter to allow the use of  ordinary functions as a ErrorHandler.
// If f is a function with the appropriate signature, ErrorHandlerFunc(f) is a ErrorHandler that calls f.
type ErrorHandlerFunc func(ctx context.Context, task *Task, err error)

// HandleError calls fn(ctx, task, err)
func (fn ErrorHandlerFunc) HandleError(ctx context.Context, task *Task, err error) {
	fn(ctx, task, err)
}

// RetryDelayFunc calculates the retry delay duration for a failed task given
// the retry count, error, and the task.
//
// n is the number of times the task canhas been retry.
// e is the error returned by the task handler.
// t is the task in question.
type RetryDelayFunc func(n int, e error, t *Task) time.Duration
