package errcol

import (
	"sync"

	"github.com/rs/zerolog"
)

type ErrorInfo struct {
	Error    string
	Severity zerolog.Level
	Expected bool
	Fields   map[string]any
	Message  string
}

type ErrorCollector struct {
	errors []ErrorInfo
	mu     *sync.RWMutex
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]ErrorInfo, 0),
		mu:     &sync.RWMutex{},
	}
}

// Checks if the error collector has any errors
// and returns true if so
func (e *ErrorCollector) HasErrors() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.errors) > 0
}

// Returns the highest severity error that the error
// collector has.
// If there are multiple of the same severity it will
// return the last one added
func (e *ErrorCollector) HighestSeverity() ErrorInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.errors) == 0 {
		return ErrorInfo{}
	}

	highestSeverity := ErrorInfo{}
	for i := len(e.errors) - 1; i >= 0; i-- {
		err := e.errors[i]
		if err.Severity > highestSeverity.Severity {
			highestSeverity = err
		}
	}
	return highestSeverity
}

// Add an error to the error collector
func (e *ErrorCollector) Add(err error, severity zerolog.Level, expected bool, fields map[string]any, message string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.errors = append(e.errors, ErrorInfo{
		Error:    err.Error(),
		Severity: severity,
		Expected: expected,
		Fields:   fields,
		Message:  message,
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
	return event.Interface("error", e.errors)
}
