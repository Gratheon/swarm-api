package graph

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitHiveMutation(t *testing.T) {
	t.Parallel()

	t.Run("split hive with new queen", func(t *testing.T) {
		t.Parallel()

		t.Run("creates new hive with new queen and moves frames", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
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

			// ACT
			newHive, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), &newQueenName, "new_queen", frameIDs)

			// ASSERT
			assert.NoError(t, err, "SplitHive failed")
			assert.NotNil(t, newHive, "Expected new hive to be created")

			var sourceQueenCount int
			db.Get(&sourceQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", sourceHiveID)
			assert.Equal(t, 1, sourceQueenCount, "Source hive should still have 1 queen")

			var sourceQueenID int
			db.Get(&sourceQueenID, "SELECT id FROM families WHERE hive_id=?", sourceHiveID)
			assert.Equal(t, queenID, sourceQueenID, "Source hive should have original queen")

			newHiveIDInt, _ := strconv.Atoi(newHive.ID)
			var newHiveQueenCount int
			db.Get(&newHiveQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", newHiveIDInt)
			assert.Equal(t, 1, newHiveQueenCount, "New hive should have 1 queen")

			var newHiveQueenName string
			db.Get(&newHiveQueenName, "SELECT name FROM families WHERE hive_id=?", newHiveIDInt)
			assert.Equal(t, newQueenName, newHiveQueenName, "New hive queen should have specified name")

			var movedFrameCount int
			db.Get(&movedFrameCount, "SELECT COUNT(*) FROM frames f JOIN boxes b ON f.box_id = b.id WHERE b.hive_id=? AND f.user_id=?", newHiveIDInt, userID)
			assert.Equal(t, 5, movedFrameCount, "Expected 5 frames to be moved to new hive")
		})
	})

	t.Run("split hive by taking old queen", func(t *testing.T) {
		t.Parallel()

		t.Run("moves existing queen to new hive leaving source queenless", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
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

			// ACT
			newHive, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), nil, "take_old_queen", frameIDs)

			// ASSERT
			assert.NoError(t, err, "SplitHive failed")
			assert.NotNil(t, newHive, "Expected new hive to be created")

			var sourceQueenCount int
			db.Get(&sourceQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", sourceHiveID)
			assert.Equal(t, 0, sourceQueenCount, "Source hive should have no queens")

			newHiveIDInt, _ := strconv.Atoi(newHive.ID)
			var newHiveQueenCount int
			db.Get(&newHiveQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", newHiveIDInt)
			assert.Equal(t, 1, newHiveQueenCount, "New hive should have 1 queen")

			var newHiveQueenID int
			db.Get(&newHiveQueenID, "SELECT id FROM families WHERE hive_id=?", newHiveIDInt)
			assert.Equal(t, queenID, newHiveQueenID, "New hive should have the old queen")
		})

		t.Run("returns error when source hive has no queen", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
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

			// ACT
			_, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), nil, "take_old_queen", frameIDs)

			// ASSERT
			assert.Error(t, err, "Expected error when trying to take queen from queenless hive")
			assert.ErrorContains(t, err, "source hive has no queen to take")
		})
	})

	t.Run("split hive without queen", func(t *testing.T) {
		t.Parallel()

		t.Run("creates queenless hive leaving source queen intact", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
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

			// ACT
			newHive, err := resolver.SplitHive(ctx, strconv.Itoa(sourceHiveID), nil, "no_queen", frameIDs)

			// ASSERT
			assert.NoError(t, err, "SplitHive failed")
			assert.NotNil(t, newHive, "Expected new hive to be created")

			var sourceQueenCount int
			db.Get(&sourceQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", sourceHiveID)
			assert.Equal(t, 1, sourceQueenCount, "Source hive should still have 1 queen")

			var sourceQueenID int
			db.Get(&sourceQueenID, "SELECT id FROM families WHERE hive_id=?", sourceHiveID)
			assert.Equal(t, queenID, sourceQueenID, "Source hive should have original queen")

			newHiveIDInt, _ := strconv.Atoi(newHive.ID)
			var newHiveQueenCount int
			db.Get(&newHiveQueenCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", newHiveIDInt)
			assert.Equal(t, 0, newHiveQueenCount, "New hive should have no queens")
		})
	})
}
