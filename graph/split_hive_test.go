package graph

import (
	"context"
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
	db.Exec("DELETE FROM boxes WHERE user_id=?", userID)
	db.Exec("DELETE FROM families WHERE user_id=?", userID)
	db.Exec("DELETE FROM hives WHERE user_id=?", userID)
	db.Exec("DELETE FROM apiaries WHERE user_id=?", userID)
}

func TestSplitHive_NewQueen(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userID := "999001"
	defer cleanupTestData(t, db, userID)

	apiaryID := createTestApiary(t, db, userID)
	sourceHiveID := createTestHive(t, db, userID, apiaryID)
	queenID := createTestQueen(t, db, userID, sourceHiveID)
	boxID := createTestBox(t, db, userID, sourceHiveID)
	frameIDs := createTestFrames(t, db, userID, boxID, 5)

	resolver := &mutationResolver{
		Resolver: &Resolver{Db: db},
	}

	ctx := context.WithValue(context.Background(), "userID", userID)

	newQueenName := "Test New Queen"
	newHive, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), &newQueenName, "new_queen", frameIDs)

	if err != nil {
		t.Fatalf("SplitHive failed: %v", err)
	}

	if newHive == nil {
		t.Fatal("Expected new hive to be created")
	}

	var sourceQueenCount int
	db.Get(&sourceQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", sourceHiveID)
	if sourceQueenCount != 1 {
		t.Errorf("Expected source hive to still have 1 queen, got %d", sourceQueenCount)
	}

	var sourceQueenID int
	db.Get(&sourceQueenID, "SELECT id FROM families WHERE hive_id=?", sourceHiveID)
	if sourceQueenID != queenID {
		t.Errorf("Expected source hive to have original queen %d, got %d", queenID, sourceQueenID)
	}

	newHiveIDInt, _ := strconv.Atoi(newHive.ID)
	var newHiveQueenCount int
	db.Get(&newHiveQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", newHiveIDInt)
	if newHiveQueenCount != 1 {
		t.Errorf("Expected new hive to have 1 queen, got %d", newHiveQueenCount)
	}

	var newHiveQueenName string
	db.Get(&newHiveQueenName, "SELECT name FROM families WHERE hive_id=?", newHiveIDInt)
	if newHiveQueenName != newQueenName {
		t.Errorf("Expected new hive queen name to be '%s', got '%s'", newQueenName, newHiveQueenName)
	}

	var movedFrameCount int
	db.Get(&movedFrameCount, "SELECT COUNT(*) FROM frames f JOIN boxes b ON f.box_id = b.id WHERE b.hive_id=? AND f.user_id=?", newHiveIDInt, userID)
	if movedFrameCount != 5 {
		t.Errorf("Expected 5 frames to be moved to new hive, got %d", movedFrameCount)
	}
}

func TestSplitHive_TakeOldQueen(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userID := "999002"
	defer cleanupTestData(t, db, userID)

	apiaryID := createTestApiary(t, db, userID)
	sourceHiveID := createTestHive(t, db, userID, apiaryID)
	queenID := createTestQueen(t, db, userID, sourceHiveID)
	boxID := createTestBox(t, db, userID, sourceHiveID)
	frameIDs := createTestFrames(t, db, userID, boxID, 3)

	resolver := &mutationResolver{
		Resolver: &Resolver{Db: db},
	}

	ctx := context.WithValue(context.Background(), "userID", userID)

	newHive, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), nil, "take_old_queen", frameIDs)

	if err != nil {
		t.Fatalf("SplitHive failed: %v", err)
	}

	if newHive == nil {
		t.Fatal("Expected new hive to be created")
	}

	var sourceQueenCount int
	db.Get(&sourceQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", sourceHiveID)
	if sourceQueenCount != 0 {
		t.Errorf("Expected source hive to have no queens, got %d", sourceQueenCount)
	}

	newHiveIDInt, _ := strconv.Atoi(newHive.ID)
	var newHiveQueenCount int
	db.Get(&newHiveQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", newHiveIDInt)
	if newHiveQueenCount != 1 {
		t.Errorf("Expected new hive to have 1 queen, got %d", newHiveQueenCount)
	}

	var newHiveQueenID int
	db.Get(&newHiveQueenID, "SELECT id FROM families WHERE hive_id=?", newHiveIDInt)
	if newHiveQueenID != queenID {
		t.Errorf("Expected new hive to have old queen %d, got %d", queenID, newHiveQueenID)
	}
}

func TestSplitHive_NoQueen(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userID := "999003"
	defer cleanupTestData(t, db, userID)

	apiaryID := createTestApiary(t, db, userID)
	sourceHiveID := createTestHive(t, db, userID, apiaryID)
	queenID := createTestQueen(t, db, userID, sourceHiveID)
	boxID := createTestBox(t, db, userID, sourceHiveID)
	frameIDs := createTestFrames(t, db, userID, boxID, 2)

	resolver := &mutationResolver{
		Resolver: &Resolver{Db: db},
	}

	ctx := context.WithValue(context.Background(), "userID", userID)

	newHive, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), nil, "no_queen", frameIDs)

	if err != nil {
		t.Fatalf("SplitHive failed: %v", err)
	}

	if newHive == nil {
		t.Fatal("Expected new hive to be created")
	}

	var sourceQueenCount int
	db.Get(&sourceQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", sourceHiveID)
	if sourceQueenCount != 1 {
		t.Errorf("Expected source hive to still have 1 queen, got %d", sourceQueenCount)
	}

	var sourceQueenID int
	db.Get(&sourceQueenID, "SELECT id FROM families WHERE hive_id=?", sourceHiveID)
	if sourceQueenID != queenID {
		t.Errorf("Expected source hive to have original queen %d, got %d", queenID, sourceQueenID)
	}

	newHiveIDInt, _ := strconv.Atoi(newHive.ID)
	var newHiveQueenCount int
	db.Get(&newHiveQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", newHiveIDInt)
	if newHiveQueenCount != 0 {
		t.Errorf("Expected new hive to have no queens, got %d", newHiveQueenCount)
	}
}

func TestSplitHive_TakeOldQueen_NoQueenInSource(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userID := "999004"
	defer cleanupTestData(t, db, userID)

	apiaryID := createTestApiary(t, db, userID)
	sourceHiveID := createTestHive(t, db, userID, apiaryID)
	boxID := createTestBox(t, db, userID, sourceHiveID)
	frameIDs := createTestFrames(t, db, userID, boxID, 2)

	resolver := &mutationResolver{
		Resolver: &Resolver{Db: db},
	}

	ctx := context.WithValue(context.Background(), "userID", userID)

	_, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), nil, "take_old_queen", frameIDs)

	if err == nil {
		t.Fatal("Expected error when trying to take queen from queenless hive")
	}

	if err.Error() != "source hive has no queen to take" {
		t.Errorf("Expected 'source hive has no queen to take' error, got: %v", err)
	}
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
