# errcol - A wide-event inspired error collector

[![Go Reference](https://pkg.go.dev/badge/github.com/daylamtayari/cierge/errcol.svg)](https://pkg.go.dev/github.com/daylamtayari/cierge/errcol)

Inspired by wide-event logging, I wanted to be able to collect all errors that occurred in a request flow to then output them at the conclusion of the event.

A core goal of this package is having errors that tell me exactly what went wrong, where, was this expected (i.e. a malformed input), and the context around the error. Unexpected errors are items that should warrant attention.

## Design

The error collector contains a slice of errors, a mutex for locking, and a boolean representing the file and line information where an error was generated should be included.

Each error contains the following attribute:
- `Error` - String value of the collected error
- `Severity` - Severity of an error, using Zerolog's levels
- `Expected` - Boolean indicating whether an error is expected (e.g. malformed user input, invalid ID, etc.) or not (issue with the server).
- `Fields` - Map that can contain additional context
- `Message` - String representing the error (e.g. "failed to retrieve object")
- `CallerInfo` - Optional attribute that includes the file path and line from where the error was called

## Usage

For every event, a new error collector should be created and then passed through the event lifecycle. I recommend using a context to store the error collector.

When an error occurs, it should be added in the following way:
```go
// Assume the errCol variable represents an ErrorCollector
errCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to perform foo")
```

At the conclusion of an event, apply the error collector to a Zerolog event:
```go
logEvent := errCol.ApplyToEvent(logEvent)
```

### Example

```go
// Assume that the ok value is always true
errCol, _ := ctx.Value(errorCollectorKey).(*errcol.ErrorCollector)

// Handling expected errors
var example any
err := json.Unmarshal(input, &example)
if err != nil {
    errCol.Add(err, zerolog.InfoLevel, true, nil, "input value is in invalid format")
}

var inputText string
if len(inputText) > 128 {
    errCol.Add(nil, zerolog.InfoLevel, true, map[string]any{"input_text": inputText}, "input text length is greater than 128")
}

// Handling unexpected errors
err := UpdateActionInDB(actionId)
if err != nil {
    errCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"action_id": actionId}, "failed to update action in database")
}

err := generateSecureHash(plaintext)
if err != nil {
    errCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to generate secure hash")
}
```

Feel free to look throughout the Cierge server package for examples, specifically including the logger middleware.
