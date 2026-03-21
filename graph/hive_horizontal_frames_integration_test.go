//go:build integration
// +build integration

package graph

import (
	"context"
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/require"
)

func TestAddHorizontalHiveCreatesOneBasedFramePositions(t *testing.T) {
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
	hiveType := model.HiveTypeHorizontal
	hiveInput := model.HiveInput{
		ApiaryID:   strconv.Itoa(apiaryID),
		BoxCount:   1,
		FrameCount: 6,
		HiveType:   &hiveType,
	}

	createdHive, err := resolver.AddHive(ctx, hiveInput)
	require.NoError(t, err)
	require.NotNil(t, createdHive)

	type frameRow struct {
		Position int `db:"position"`
	}

	rows := make([]frameRow, 0)
	err = db.Select(
		&rows,
		`SELECT f.position
		FROM frames f
		INNER JOIN boxes b ON b.id = f.box_id
		WHERE f.user_id = ? AND b.hive_id = ? AND b.type = 'LARGE_HORIZONTAL_SECTION' AND f.active = 1
		ORDER BY f.position ASC`,
		userID,
		createdHive.ID,
	)
	require.NoError(t, err)
	require.Len(t, rows, 6)

	for i, row := range rows {
		require.Equal(t, i+1, row.Position)
	}
}
