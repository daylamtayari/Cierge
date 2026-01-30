package errcol

import (
	"runtime"
	"sync"

	"github.com/rs/zerolog"
)

// Stores the information on the
// caller, getting the last frame.
// The File is the full path of
// the file that called the error
// and Line is the line number
// that the error was added from
type CallerInfo struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// Stores information about an error
type ErrorInfo struct {
	Error      string         `json:"error"`
	Severity   zerolog.Level  `json:"severity"`
	Expected   bool           `json:"expected"`
	Fields     map[string]any `json:"fields,omitempty"`
	Message    string         `json:"message"`
	CallerInfo CallerInfo     `json:"caller_info"`
}

// An error collector that stores
// a list of errors
type ErrorCollector struct {
	errors     []ErrorInfo
	mu         *sync.RWMutex
	callerInfo bool
}

// Creates a new error collector
// If callerInfo is set to true, the file and line information
// of where an error was called is stored with an error
func NewErrorCollector(callerInfo bool) *ErrorCollector {
	return &ErrorCollector{
		errors:     make([]ErrorInfo, 0),
		mu:         &sync.RWMutex{},
		callerInfo: callerInfo,
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

	errorInfo := ErrorInfo{
		Error:    err.Error(),
		Severity: severity,
		Expected: expected,
		Fields:   fields,
		Message:  message,
	}

	if e.callerInfo {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			errorInfo.CallerInfo = CallerInfo{
				File: file,
				Line: line,
			}
		}
	}

	e.errors = append(e.errors, errorInfo)
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
