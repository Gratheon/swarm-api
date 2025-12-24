package graph

import (
	"os"
	"strconv"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:test@tcp(localhost:5100)/swarm-api?parseTime=true"
	}

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to test database: %v\nMake sure mysql is running: cd ../mysql && just start", err)
		return nil
	}
	return db
}

func cleanupTestData(t *testing.T, db *sqlx.DB, userID string) {
	db.Exec("DELETE FROM frames WHERE user_id=?", userID)
	db.Exec("DELETE FROM frames_sides WHERE user_id=?", userID)
	db.Exec("DELETE FROM boxes WHERE user_id=?", userID)
	db.Exec("DELETE FROM families WHERE user_id=?", userID)
	db.Exec("DELETE FROM hives WHERE user_id=?", userID)
	db.Exec("DELETE FROM apiaries WHERE user_id=?", userID)
}

func createTestApiary(t *testing.T, db *sqlx.DB, userID string) int {
	result := db.MustExec(
		"INSERT INTO apiaries (user_id, name, active) VALUES (?, 'Test Apiary', 1)",
		userID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestHive(t *testing.T, db *sqlx.DB, userID string, apiaryID int) int {
	result := db.MustExec(
		"INSERT INTO hives (user_id, apiary_id, active, added) VALUES (?, ?, 1, NOW())",
		userID, apiaryID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestQueen(t *testing.T, db *sqlx.DB, userID string, hiveID int) int {
	queenName := "Test Queen"
	result := db.MustExec(
		"INSERT INTO families (user_id, hive_id, name) VALUES (?, ?, ?)",
		userID, hiveID, queenName,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestBox(t *testing.T, db *sqlx.DB, userID string, hiveID int) int {
	result := db.MustExec(
		"INSERT INTO boxes (user_id, hive_id, position, type, active) VALUES (?, ?, 0, 'DEEP', 1)",
		userID, hiveID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestFrames(t *testing.T, db *sqlx.DB, userID string, boxID int, count int) []string {
	frameIDs := make([]string, count)
	for i := 0; i < count; i++ {
		result := db.MustExec(
			"INSERT INTO frames (user_id, box_id, position, type, active) VALUES (?, ?, ?, 'EMPTY_COMB', 1)",
			userID, boxID, i,
		)
		id, _ := result.LastInsertId()
		frameIDs[i] = strconv.Itoa(int(id))
	}
	return frameIDs
}

func createTestFrameSide(t *testing.T, db *sqlx.DB, userID string) int {
	result := db.MustExec(
		"INSERT INTO frames_sides (user_id) VALUES (?)",
		userID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestFrameWithSides(t *testing.T, db *sqlx.DB, userID string, boxID int, position int, leftSideID int, rightSideID int) int {
	result := db.MustExec(
		"INSERT INTO frames (user_id, box_id, position, type, active, left_id, right_id) VALUES (?, ?, ?, 'EMPTY_COMB', 1, ?, ?)",
		userID, boxID, position, leftSideID, rightSideID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}
