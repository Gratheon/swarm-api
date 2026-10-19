//go:build integration
// +build integration

package graph

import (
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
)

var integrationTestUserIDCounter uint64 = 900000000

func createTestUserID() string {
	return strconv.FormatUint(atomic.AddUint64(&integrationTestUserIDCounter, 1), 10)
}

func setupTestDB(t *testing.T) *sqlx.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@localhost:5432/swarm_api?sslmode=disable&connect_timeout=5"
	}

	// Same driver setup as the service (placeholders, UTC session, legacy date format)
	db, err := OpenDB(dsn)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to test database: %v\nMake sure postgres is running and migrated: just migrate-db-dev", err)
		return nil
	}

	return db
}

func cleanupTestData(t *testing.T, db *sqlx.DB, userID string) {
	db.Exec("DELETE FROM hive_logs WHERE user_id=$1", userID)
	db.Exec("DELETE FROM inspections WHERE user_id=$1", userID)
	db.Exec("DELETE FROM treatments WHERE user_id=$1", userID)
	db.Exec("DELETE FROM family_moves WHERE user_id=$1", userID)
	db.Exec("DELETE FROM frames WHERE user_id=$1", userID)
	db.Exec("DELETE FROM frames_sides WHERE user_id=$1", userID)
	db.Exec("DELETE FROM boxes WHERE user_id=$1", userID)
	db.Exec("DELETE FROM families WHERE user_id=$1", userID)
	db.Exec("DELETE FROM hives WHERE user_id=$1", userID)
	db.Exec("DELETE FROM apiaries WHERE user_id=$1", userID)
}

func createTestApiary(t *testing.T, db *sqlx.DB, userID string) int {
	var id int64
	if err := db.QueryRowx(
		"INSERT INTO apiaries (user_id, name, active) VALUES ($1, 'Test Apiary', TRUE) RETURNING id",
		userID,
	).Scan(&id); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	return int(id)
}

func createTestHive(t *testing.T, db *sqlx.DB, userID string, apiaryID int) int {
	var id int64
	if err := db.QueryRowx(
		"INSERT INTO hives (user_id, apiary_id, active, added) VALUES ($1, $2, TRUE, NOW()) RETURNING id",
		userID, apiaryID,
	).Scan(&id); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	return int(id)
}

func createTestQueen(t *testing.T, db *sqlx.DB, userID string, hiveID int) int {
	queenName := "Test Queen"
	var id int64
	if err := db.QueryRowx(
		"INSERT INTO families (user_id, hive_id, name) VALUES ($1, $2, $3) RETURNING id",
		userID, hiveID, queenName,
	).Scan(&id); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	return int(id)
}

func createTestBox(t *testing.T, db *sqlx.DB, userID string, hiveID int) int {
	var id int64
	if err := db.QueryRowx(
		"INSERT INTO boxes (user_id, hive_id, position, type, active) VALUES ($1, $2, 0, 'DEEP', TRUE) RETURNING id",
		userID, hiveID,
	).Scan(&id); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	return int(id)
}

func createTestFrames(t *testing.T, db *sqlx.DB, userID string, boxID int, count int) []string {
	frameIDs := make([]string, count)
	for i := 0; i < count; i++ {
		var id int64
		if err := db.QueryRowx(
			"INSERT INTO frames (user_id, box_id, position, type, active) VALUES ($1, $2, $3, 'EMPTY_COMB', TRUE) RETURNING id",
			userID, boxID, i,
		).Scan(&id); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
		frameIDs[i] = strconv.Itoa(int(id))
	}
	return frameIDs
}

func createTestFrameSide(t *testing.T, db *sqlx.DB, userID string) int {
	var id int64
	if err := db.QueryRowx(
		"INSERT INTO frames_sides (user_id) VALUES ($1) RETURNING id",
		userID,
	).Scan(&id); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	return int(id)
}

func createTestFrameWithSides(t *testing.T, db *sqlx.DB, userID string, boxID int, position int, leftSideID int, rightSideID int) int {
	var id int64
	if err := db.QueryRowx(
		"INSERT INTO frames (user_id, box_id, position, type, active, left_id, right_id) VALUES ($1, $2, $3, 'EMPTY_COMB', TRUE, $4, $5) RETURNING id",
		userID, boxID, position, leftSideID, rightSideID,
	).Scan(&id); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	return int(id)
}
