package graph

import (
	"context"
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestApiaryWithActiveFlag(t *testing.T, db *sqlx.DB, userID string, active bool) int {
	activeValue := 0
	if active {
		activeValue = 1
	}

	result := db.MustExec(
		"INSERT INTO apiaries (user_id, name, active) VALUES (?, 'Test Apiary', ?)",
		userID,
		activeValue,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func TestHiveCreationLimit(t *testing.T) {
	t.Parallel()

	t.Run("free tier allows third hive and blocks fourth", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999201"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		for i := 0; i < 2; i++ {
			createTestHive(t, db, userID, apiaryID)
		}

		resolver := &mutationResolver{
			Resolver: &Resolver{Db: db},
		}
		ctx := context.WithValue(context.Background(), "userID", userID)
		ctx = context.WithValue(ctx, "billingPlan", "free")

		hiveInput := model.HiveInput{
			ApiaryID:   strconv.Itoa(apiaryID),
			BoxCount:   1,
			FrameCount: 1,
		}

		// ACT
		hiveAtLimit, errAtLimit := resolver.AddHive(ctx, hiveInput)
		hiveOverLimit, errOverLimit := resolver.AddHive(ctx, hiveInput)

		// ASSERT
		require.NoError(t, errAtLimit, "Expected to be able to create hive at the exact free tier limit")
		require.NotNil(t, hiveAtLimit, "Expected created hive at free tier limit")
		assert.Error(t, errOverLimit, "Expected hive creation above free tier limit to fail")
		assert.ErrorContains(t, errOverLimit, "hive limit reached for free plan (3)")
		assert.Nil(t, hiveOverLimit, "Expected no hive to be returned when over free tier limit")
	})

	t.Run("hobbyist tier allows fifteenth hive and blocks sixteenth", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999202"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		for i := 0; i < 14; i++ {
			createTestHive(t, db, userID, apiaryID)
		}

		resolver := &mutationResolver{
			Resolver: &Resolver{Db: db},
		}
		ctx := context.WithValue(context.Background(), "userID", userID)
		ctx = context.WithValue(ctx, "billingPlan", "hobbyist")

		hiveInput := model.HiveInput{
			ApiaryID:   strconv.Itoa(apiaryID),
			BoxCount:   1,
			FrameCount: 1,
		}

		// ACT
		hiveAtLimit, errAtLimit := resolver.AddHive(ctx, hiveInput)
		hiveOverLimit, errOverLimit := resolver.AddHive(ctx, hiveInput)

		// ASSERT
		require.NoError(t, errAtLimit, "Expected to be able to create hive at the exact hobbyist tier limit")
		require.NotNil(t, hiveAtLimit, "Expected created hive at hobbyist tier limit")
		assert.Error(t, errOverLimit, "Expected hive creation above hobbyist tier limit to fail")
		assert.ErrorContains(t, errOverLimit, "hive limit reached for hobbyist plan (15)")
		assert.Nil(t, hiveOverLimit, "Expected no hive to be returned when over hobbyist tier limit")
	})

	t.Run("limit count ignores hives in deactivated apiaries", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999203"
		defer cleanupTestData(t, db, userID)

		activeApiaryID := createTestApiary(t, db, userID)
		inactiveApiaryID := createTestApiaryWithActiveFlag(t, db, userID, false)

		for i := 0; i < 14; i++ {
			createTestHive(t, db, userID, activeApiaryID)
		}
		for i := 0; i < 20; i++ {
			createTestHive(t, db, userID, inactiveApiaryID)
		}

		resolver := &mutationResolver{
			Resolver: &Resolver{Db: db},
		}
		ctx := context.WithValue(context.Background(), "userID", userID)
		ctx = context.WithValue(ctx, "billingPlan", "hobbyist")

		hiveInput := model.HiveInput{
			ApiaryID:   strconv.Itoa(activeApiaryID),
			BoxCount:   1,
			FrameCount: 1,
		}

		// ACT
		createdHive, err := resolver.AddHive(ctx, hiveInput)

		// ASSERT
		require.NoError(t, err, "Expected hives in deactivated apiaries to be ignored for limit checks")
		require.NotNil(t, createdHive, "Expected hive creation to succeed when active apiary hive count is under limit")
	})
}
