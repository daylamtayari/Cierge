package errcol

import (
	"sync"

	"github.com/rs/zerolog"
)

type ErrorInfo struct {
	Message  string
	Severity zerolog.Level
	Expected bool
}

type ErrorCollector struct {
	errors []ErrorInfo
	mu     *sync.RWMutex
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]ErrorInfo, 0),
	}
}

// Checks if the error collector has any errors
// and returns true if so
func (e *ErrorCollector) HasErrors() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.errors) > 0
}

// Add an error to the error collector
func (e *ErrorCollector) Add(err error, severity zerolog.Level, expected bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.errors = append(e.errors, ErrorInfo{
		Message:  err.Error(),
		Severity: severity,
		Expected: expected,
	})
}

// Applies all of the errors to a given zerolog event, setting them all
// to the errors field
func (e *ErrorCollector) ApplyToEvent(event *zerolog.Event) *zerolog.Event {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.errors) == 0 {
		return event
	}
	return event.Interface("errors", e.errors)
}
