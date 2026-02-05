package querycol

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	versionInfo map[string]string
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

// Set version information that you care to track
// e.g. database version, sql driver version, gorm version
// Allows to have greater insight and analysis towards
// DB performance and identify anomalies that can be
// related to version changes.
func SetVersions(versions map[string]string) {
	versionInfo = versions
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
	}

	if err != nil {
		query.Error = err.Error()
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
		TotalQueries  int               `json:"total_queries"`
		TotalDuration time.Duration     `json:"total_duration"`
		Versions      map[string]string `json:"version,omitempty"`
		Queries       []QueryInfo       `json:"queries"`
	}{
		TotalQueries:  len(q.queries),
		TotalDuration: q.TotalDuration(),
		Versions:      versionInfo,
		Queries:       q.queries,
	}

	return event.Interface("database", queryOutput)
}
