# querycol - A wide-event inspired database query collector

Inspired by wide-event logging, I wanted to be able to collect all database queries that occurred in a request flow, and then output them at the conclusion of the event.

## Design

The query collector contains a slice of queries, a mutex for locking, and a boolean representing whether the SQL of the queries should be included. SQL of queries should not be included in production as it can disclose sensitive data.

Each query contains the following attributes:
- `SQL` - String value representing the raw SQL query made
- `Duration` - Duration of the query
- `Rows` - The number of rows the query affected
- `Error` - Error message if the query encountered an error

The following attributes are returned when applying the query collector to an event:
- `TotalQueries` - Number of queries performed
- `TotalDuration` - Total duration of all queries
- `Versions` - String map containing version information, set when creating a new query collector (I use it to track PostgreSQL, gorm, and gorm driver versions)
- `Queries` - Slice of queries with the above attributes

## Usage

For every event, a new query collector should be created and then passed through the event lifecycle. I recommend using a context to store the query collector.

The logger interface of the database client should be overridden to add queries to the query collector.

At the conclusion of an event, apply the query collector to the Zerolog event:
```go
logEvent = queryCol.ApplyToEvent(logEvent)
```

### Example Implementation for `gorm.DB`

```go
type Logger struct {
	logLevel gormlogger.LogLevel
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Add query to the query collector
    // appctx.QueryCollector is a wrapper method to get the query collector from the context
	appctx.QueryCollector(ctx).Add(sql, elapsed, rows, err)
}
```
