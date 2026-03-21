package graph

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/qustavo/sqlhooks/v2"
)

const (
	instrumentedMySQLDriverName = "mysql+metrics"
	maxDBQueryLabelCardinality  = 200
)

var dbMetricsRegisterer = prometheus.WrapRegistererWith(
	prometheus.Labels{"service": "swarm-api"},
	prometheus.DefaultRegisterer,
)

var dbQueriesTotal = promauto.With(dbMetricsRegisterer).NewCounterVec(
	prometheus.CounterOpts{
		Name: "swarm_api_db_queries_total",
		Help: "Total number of DB queries",
	},
	[]string{"operation", "query", "status"},
)

var dbQueryDurationSeconds = promauto.With(dbMetricsRegisterer).NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "swarm_api_db_query_duration_seconds",
		Help:    "DB query duration in seconds",
		Buckets: []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	},
	[]string{"operation", "query", "status"},
)

var dbQueryStateKey = struct{}{}

type dbQueryState struct {
	start  time.Time
	failed bool
}

type dbQueryMetricsHooks struct{}

var (
	singleQuotedStringRegex = regexp.MustCompile(`'([^'\\]|\\.)*'`)
	doubleQuotedStringRegex = regexp.MustCompile(`"([^"\\]|\\.)*"`)
	numberRegex             = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	inListRegex             = regexp.MustCompile(`(?i)\bin\s*\((\s*\?\s*,?)+\)`)
	whitespaceRegex         = regexp.MustCompile(`\s+`)
)

var (
	registerInstrumentedDriverOnce sync.Once
	queryLabelCardinalityMutex     sync.Mutex
	knownDBQueryLabels             = map[string]struct{}{}
)

func ensureInstrumentedMySQLDriverRegistered() {
	registerInstrumentedDriverOnce.Do(func() {
		sql.Register(instrumentedMySQLDriverName, sqlhooks.Wrap(&mysqlDriver.MySQLDriver{}, &dbQueryMetricsHooks{}))
	})
}

func (h *dbQueryMetricsHooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, dbQueryStateKey, &dbQueryState{
		start: time.Now(),
	}), nil
}

func (h *dbQueryMetricsHooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	state, _ := ctx.Value(dbQueryStateKey).(*dbQueryState)
	if state == nil {
		state = &dbQueryState{start: time.Now()}
	}

	status := "success"
	if state.failed {
		status = "error"
	}

	observeDBQueryMetrics(query, status, time.Since(state.start))
	return ctx, nil
}

func (h *dbQueryMetricsHooks) OnError(ctx context.Context, _ error, _ string, _ ...interface{}) error {
	if state, ok := ctx.Value(dbQueryStateKey).(*dbQueryState); ok && state != nil {
		state.failed = true
	}
	return nil
}

func observeDBQueryMetrics(query string, status string, duration time.Duration) {
	operation := extractOperation(query)
	queryLabel := boundQueryLabelCardinality(normalizeQuery(query))

	dbQueriesTotal.WithLabelValues(operation, queryLabel, status).Inc()
	dbQueryDurationSeconds.WithLabelValues(operation, queryLabel, status).Observe(duration.Seconds())
}

func extractOperation(query string) string {
	trimmed := strings.TrimSpace(strings.ToLower(query))
	if trimmed == "" {
		return "unknown"
	}

	firstToken := strings.Fields(trimmed)
	if len(firstToken) == 0 {
		return "unknown"
	}

	switch firstToken[0] {
	case "select", "insert", "update", "delete", "replace", "upsert", "with":
		return firstToken[0]
	default:
		return "other"
	}
}

func normalizeQuery(query string) string {
	normalized := strings.TrimSpace(query)
	if normalized == "" {
		return "unknown"
	}

	normalized = strings.ToLower(normalized)
	normalized = singleQuotedStringRegex.ReplaceAllString(normalized, "?")
	normalized = doubleQuotedStringRegex.ReplaceAllString(normalized, "?")
	normalized = numberRegex.ReplaceAllString(normalized, "?")
	normalized = inListRegex.ReplaceAllString(normalized, "in (?)")
	normalized = whitespaceRegex.ReplaceAllString(normalized, " ")

	if len(normalized) > 160 {
		normalized = normalized[:160]
	}

	return normalized
}

func boundQueryLabelCardinality(query string) string {
	queryLabelCardinalityMutex.Lock()
	defer queryLabelCardinalityMutex.Unlock()

	if _, exists := knownDBQueryLabels[query]; exists {
		return query
	}
	if len(knownDBQueryLabels) >= maxDBQueryLabelCardinality {
		return "__other__"
	}

	knownDBQueryLabels[query] = struct{}{}
	return query
}
