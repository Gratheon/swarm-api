//go:build integration
// +build integration

package graph

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// Fills every connection slot the test role can use, frees one shortly after, and checks
// that a new pool connection waits for it instead of failing with SQLSTATE 53300.
func TestOpenDB_RetriesWhenConnectionSlotsAreFull(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@localhost:5432/swarm_api?sslmode=disable&connect_timeout=5"
	}
	db := setupTestDB(t)
	defer db.Close()
	db.SetMaxOpenConns(1)
	require.NoError(t, db.Ping())
	db.SetMaxIdleConns(0) // the next query has to open a new connection

	ctx := context.Background()
	var fillers []*pgx.Conn
	defer func() {
		for _, c := range fillers {
			c.Close(ctx)
		}
	}()
	for i := 0; i < 1000; i++ {
		c, err := pgx.Connect(ctx, dsn)
		if err != nil {
			if !isTooManyConnections(err) {
				// e.g. PgBouncer's max_client_conn: slots are only exhausted on direct Postgres
				t.Skipf("server refused with a different error, not testing slot exhaustion: %v", err)
			}
			break
		}
		fillers = append(fillers, c)
	}
	require.NotEmpty(t, fillers)

	last := fillers[len(fillers)-1]
	fillers = fillers[:len(fillers)-1]
	go func() {
		time.Sleep(150 * time.Millisecond)
		last.Close(ctx)
	}()

	var one int
	require.NoError(t, db.Get(&one, "SELECT 1"))
	require.Equal(t, 1, one)
}
