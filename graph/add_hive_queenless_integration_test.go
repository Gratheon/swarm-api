//go:build integration
// +build integration

package graph

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/require"
)

func TestAddHiveCreatesNoInitialFamilyWithoutLegacyQueenName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userID := createTestUserID()
	defer cleanupTestData(t, db, userID)

	apiaryID := createTestApiary(t, db, userID)
	resolver := &mutationResolver{
		Resolver: &Resolver{Db: db},
	}

	ctx := context.WithValue(context.Background(), "userID", userID)
	hiveInput := model.HiveInput{
		ApiaryID:   strconv.Itoa(apiaryID),
		BoxCount:   1,
		FrameCount: 10,
	}

	createdHive, err := resolver.AddHive(ctx, hiveInput)
	require.NoError(t, err)
	require.NotNil(t, createdHive)

	var familyCount int
	err = db.Get(&familyCount, "SELECT COUNT(*) FROM families WHERE hive_id=?", createdHive.ID)
	require.NoError(t, err)
	require.Equal(t, 0, familyCount)
}

func TestAddHiveCreatesInitialFamilyForLegacyQueenName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	userID := createTestUserID()
	defer cleanupTestData(t, db, userID)

	apiaryID := createTestApiary(t, db, userID)
	resolver := &mutationResolver{
		Resolver: &Resolver{Db: db},
	}

	ctx := context.WithValue(context.Background(), "userID", userID)
	queenName := "  Legacy Queen  "
	queenYear := "2025"
	queenColor := "blue"
	hiveInput := model.HiveInput{
		ApiaryID:   strconv.Itoa(apiaryID),
		BoxCount:   1,
		FrameCount: 10,
		QueenName:  &queenName,
		QueenYear:  &queenYear,
		QueenColor: &queenColor,
	}

	createdHive, err := resolver.AddHive(ctx, hiveInput)
	require.NoError(t, err)
	require.NotNil(t, createdHive)

	type familyRow struct {
		Name  sql.NullString `db:"name"`
		Added sql.NullString `db:"added"`
		Color sql.NullString `db:"color"`
	}

	rows := make([]familyRow, 0)
	err = db.Select(&rows, "SELECT name, added, color FROM families WHERE hive_id=?", createdHive.ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "Legacy Queen", rows[0].Name.String)
	require.Equal(t, queenYear, rows[0].Added.String)
	require.Equal(t, queenColor, rows[0].Color.String)
}
