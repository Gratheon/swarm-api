//go:build integration
// +build integration

package graph

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutationHiveResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Hive", func(t *testing.T) {
		t.Parallel()

		t.Run("JoinHives", func(t *testing.T) {
			t.Parallel()

			t.Run("MergesSourceHiveAndMovesBoxes", func(t *testing.T) {
				t.Parallel()

				// ARRANGE
				fx := newSchemaResolverFixture(t, false)
				sourceApiaryID := createTestApiary(t, fx.resolver.Db, fx.userID)
				targetApiaryID := createTestApiary(t, fx.resolver.Db, fx.userID)
				sourceHiveID := createTestHive(t, fx.resolver.Db, fx.userID, sourceApiaryID)
				targetHiveID := createTestHive(t, fx.resolver.Db, fx.userID, targetApiaryID)
				sourceBoxID := createTestBox(t, fx.resolver.Db, fx.userID, sourceHiveID)
				_ = createTestBox(t, fx.resolver.Db, fx.userID, targetHiveID)
				_, sourcePlacementErr := (&model.HivePlacement{
					Db:     fx.resolver.Db,
					UserID: fx.userID,
				}).Update(strconv.Itoa(sourceApiaryID), strconv.Itoa(sourceHiveID), 1, 1, 0)
				require.NoError(t, sourcePlacementErr)

				// ACT
				mergedTargetHive, err := fx.mutation.JoinHives(
					fx.ctx,
					strconv.Itoa(sourceHiveID),
					strconv.Itoa(targetHiveID),
					"both_queens",
				)

				// ASSERT
				require.NoError(t, err)
				require.NotNil(t, mergedTargetHive)
				assert.Equal(t, strconv.Itoa(targetHiveID), mergedTargetHive.ID)
				var mergedStatus string
				var mergedInto sql.NullInt64
				statusErr := fx.resolver.Db.Get(
					&mergedStatus,
					"SELECT status FROM hives WHERE id=? AND user_id=? LIMIT 1",
					sourceHiveID,
					fx.userID,
				)
				require.NoError(t, statusErr)
				mergedIntoErr := fx.resolver.Db.Get(
					&mergedInto,
					"SELECT merged_into_hive_id FROM hives WHERE id=? AND user_id=? LIMIT 1",
					sourceHiveID,
					fx.userID,
				)
				require.NoError(t, mergedIntoErr)
				assert.Equal(t, "merged", mergedStatus)
				require.True(t, mergedInto.Valid)
				assert.Equal(t, int64(targetHiveID), mergedInto.Int64)
				var movedBoxCount int
				movedErr := fx.resolver.Db.Get(
					&movedBoxCount,
					"SELECT COUNT(*) FROM boxes WHERE id=? AND hive_id=? AND user_id=? AND active=1",
					sourceBoxID,
					targetHiveID,
					fx.userID,
				)
				require.NoError(t, movedErr)
				assert.Equal(t, 1, movedBoxCount)
				var sourcePlacementCount int
				placementErr := fx.resolver.Db.Get(
					&sourcePlacementCount,
					"SELECT COUNT(*) FROM hive_placements WHERE hive_id=? AND user_id=?",
					sourceHiveID,
					fx.userID,
				)
				require.NoError(t, placementErr)
				assert.Equal(t, 0, sourcePlacementCount)
			})

			t.Run("RejectsInvalidMergeType", func(t *testing.T) {
				t.Parallel()

				// ARRANGE
				fx := newSchemaResolverFixture(t, true)

				// ACT
				_, err := fx.mutation.JoinHives(
					fx.ctx,
					strconv.Itoa(fx.hiveID),
					strconv.Itoa(fx.hiveID),
					"invalid-merge-type",
				)

				// ASSERT
				require.Error(t, err)
				assert.ErrorContains(t, err, "invalid merge type")
			})
		})

		t.Run("UpdateHivePlacement", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			apiaryID := createTestApiary(t, fx.resolver.Db, fx.userID)
			hiveID := createTestHive(t, fx.resolver.Db, fx.userID, apiaryID)
			apiaryIDString := strconv.Itoa(apiaryID)
			hiveIDString := strconv.Itoa(hiveID)

			// ACT
			created, createErr := fx.mutation.UpdateHivePlacement(fx.ctx, apiaryIDString, hiveIDString, 1.25, 2.5, 30)
			updated, updateErr := fx.mutation.UpdateHivePlacement(fx.ctx, apiaryIDString, hiveIDString, 9.5, 8.5, 180)
			placements, listErr := fx.query.HivePlacements(fx.ctx, apiaryIDString)

			// ASSERT
			require.NoError(t, createErr)
			require.NotNil(t, created)
			require.NoError(t, updateErr)
			require.NotNil(t, updated)
			assert.Equal(t, created.ID, updated.ID)
			assert.Equal(t, 9.5, updated.X)
			assert.Equal(t, 8.5, updated.Y)
			assert.Equal(t, 180.0, updated.Rotation)
			require.NoError(t, listErr)
			assert.True(t, hasHivePlacementID(placements, updated.ID))
		})
	})
}
