package querycol

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Stores information about a
// database query
type QueryInfo struct {
	SQL      string        `json:"sql,omitempty"`
	Duration time.Duration `json:"duration"`
	Rows     int64         `json:"rows"`
	Error    string        `json:"error,omitempty"`
}

// Stores a list of database query information
type QueryCollector struct {
	queries    []QueryInfo
	mu         sync.RWMutex
	includeSql bool
}

// Creates a new query collector
// If includeSql is set to true, the SQL query is stored
// NOTE: Only set includeSql to true in a development environment,
// by nature of including SQL queries, it will disclose user information
// that should not be present in production logs
func NewQueryCollector(includeSql bool) *QueryCollector {
	return &QueryCollector{
		queries:    make([]QueryInfo, 0),
		includeSql: includeSql,
	}
}

// Returns true if any queries were stored
func (q *QueryCollector) HasQueries() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queries) > 0
}

// Returns the total number of queries collected
func (q *QueryCollector) TotalQueries() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queries)
}

// Returns the sum of all query durations
func (q *QueryCollector) TotalDuration() time.Duration {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var total time.Duration
	for _, q := range q.queries {
		total += q.Duration
	}
	return total
}

// Adds a query to the query collector
// The sql string will only be stored in the collector if
// includeSql was set to true in the constructor
func (q *QueryCollector) Add(sql string, duration time.Duration, rows int64, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	query := QueryInfo{
		Duration: duration,
		Rows:     rows,
		Error:    err.Error(),
	}

	if q.includeSql {
		query.SQL = sql
	}

	q.queries = append(q.queries, query)
}

// Applies all of the queries as well as a total queries and total query duration
// field to the database field of a zerolog event
func (q *QueryCollector) ApplyToEvent(event *zerolog.Event) *zerolog.Event {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.queries) == 0 {
		return event
	}

	queryOutput := struct {
		TotalQueries  int
		TotalDuration time.Duration
		Queries       []QueryInfo
	}{
		TotalQueries:  len(q.queries),
		TotalDuration: q.TotalDuration(),
		Queries:       q.queries,
	}

	return event.Interface("database", queryOutput)
}
