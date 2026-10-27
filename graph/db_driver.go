package graph

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/qustavo/sqlhooks/v2"
)

const instrumentedDriverName = "pgx+metrics"

// apiDateTimeLayout is the date format the API has always returned: the service ran on MySQL
// without parseTime, so DATETIME columns were scanned into Go strings as "YYYY-MM-DD HH:MM:SS".
const apiDateTimeLayout = "2006-01-02 15:04:05"

var registerInstrumentedDriverOnce sync.Once

func ensureInstrumentedDriverRegistered() {
	registerInstrumentedDriverOnce.Do(func() {
		sql.Register(instrumentedDriverName, sqlhooks.Wrap(&dateTimeTextDriver{parent: stdlib.GetDefaultDriver()}, &dbQueryMetricsHooks{}))
		// Named queries (NamedExec) and Rebind() must produce $1..$n for this custom driver name
		sqlx.BindDriver(instrumentedDriverName, sqlx.DOLLAR)
	})
}

// OpenDB connects to Postgres with query metrics and the legacy date format.
// dsn is a libpq URL or key/value string (e.g. postgres://user:pass@host:5432/db?sslmode=require).
func OpenDB(dsn string) (*sqlx.DB, error) {
	ensureInstrumentedDriverRegistered()

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	// Columns are TIMESTAMPTZ and the API exposes zone-less strings, so every session must be UTC
	// regardless of the server/database default. Also makes string date params parse as UTC.
	// "TimeZone" is the spelling PgBouncer tracks per client, so this also works through a
	// DigitalOcean connection pool in transaction mode.
	delete(config.RuntimeParams, "timezone")
	config.RuntimeParams["TimeZone"] = "UTC"
	// Named prepared statements (pgx's default statement cache) break behind PgBouncer in
	// transaction mode ("prepared statement ... does not exist"). The exec mode sends each
	// query as an unnamed statement in one round trip, so the same DSN works directly and
	// through a DigitalOcean connection pool. A default_query_exec_mode DSN param still wins.
	if !strings.Contains(dsn, "default_query_exec_mode") {
		config.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	return sqlx.Connect(instrumentedDriverName, stdlib.RegisterConnConfig(config))
}

// isTooManyConnections reports SQLSTATE 53300 ("too many clients" / "remaining connection
// slots are reserved"), a transient error when the shared cluster is at max_connections.
func isTooManyConnections(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "53300"
}

// pgxConn is the set of driver interfaces implemented by the pgx stdlib connection.
type pgxConn interface {
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.ExecerContext
	driver.QueryerContext
	driver.Pinger
	driver.NamedValueChecker
	driver.SessionResetter
}

// dateTimeTextDriver wraps the pgx stdlib driver and returns timestamp values as
// "YYYY-MM-DD HH:MM:SS" (UTC) strings instead of time.Time.
//
// WHY: all date fields of the models are Go strings. database/sql would format a time.Time
// as RFC3339 ("2026-01-02T03:04:05Z"), silently changing the public GraphQL date format
// (and breaking time.Parse(time.DateTime, ...) in resolvers). pgx stdlib hard-codes
// time.Time for timestamp OIDs, so the conversion has to happen at the driver level.
// Consequence: scanning a timestamp into time.Time / sql.NullTime is not supported.
type dateTimeTextDriver struct {
	parent driver.Driver
}

// connectRetryDelays are the waits before re-dialing when the server has no free
// connection slots. Other services sharing the cluster usually release one within a second.
var connectRetryDelays = []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, 700 * time.Millisecond}

func (d *dateTimeTextDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	for _, delay := range connectRetryDelays {
		if err == nil || !isTooManyConnections(err) {
			break
		}
		time.Sleep(delay)
		conn, err = d.parent.Open(name)
	}
	if err != nil {
		return nil, err
	}
	pc, ok := conn.(pgxConn)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("unexpected postgres driver connection type %T", conn)
	}
	return &dateTimeTextConn{pgxConn: pc}, nil
}

type dateTimeTextConn struct {
	pgxConn
}

func (c *dateTimeTextConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows, err := c.pgxConn.QueryContext(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return &dateTimeTextRows{Rows: rows}, nil
}

func (c *dateTimeTextConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *dateTimeTextConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	stmt, err := c.pgxConn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &dateTimeTextStmt{Stmt: stmt}, nil
}

type dateTimeTextStmt struct {
	driver.Stmt
}

func (s *dateTimeTextStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.Stmt.(driver.StmtExecContext).ExecContext(ctx, args)
}

func (s *dateTimeTextStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, err := s.Stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	return &dateTimeTextRows{Rows: rows}, nil
}

type dateTimeTextRows struct {
	driver.Rows
}

func (r *dateTimeTextRows) Next(dest []driver.Value) error {
	if err := r.Rows.Next(dest); err != nil {
		return err
	}
	for i, value := range dest {
		if t, ok := value.(time.Time); ok {
			dest[i] = t.UTC().Format(apiDateTimeLayout)
		}
	}
	return nil
}
