//go:build integration
// +build integration

package graph

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schemaResolverFixture struct {
	ctx      context.Context
	userID   string
	resolver *Resolver

	mutation *mutationResolver
	query    *queryResolver
	apiary   *apiaryResolver
	hive     *hiveResolver
	box      *boxResolver
	frame    *frameResolver
	family   *familyResolver

	apiaryID     int
	hiveID       int
	familyID     int
	boxID        int
	frameID      int
	leftSideID   int
	rightSideID  int
	inspectionID string
}

func newSchemaResolverFixture(t *testing.T, seed bool) *schemaResolverFixture {
	t.Helper()

	db := setupTestDB(t)
	require.NotNil(t, db, "schema resolver fixture failed to initialize test database")
	t.Cleanup(func() {
		db.Close()
	})

	userID := createTestUserID()
	t.Cleanup(func() {
		cleanupTestData(t, db, userID)
	})

	resolver := &Resolver{Db: db, femaleNamesMap: map[string][]string{"en": {"Alice", "Bella"}, "et": {"Kati"}}}
	ctx := context.WithValue(context.Background(), "userID", userID)

	fx := &schemaResolverFixture{
		ctx:      ctx,
		userID:   userID,
		resolver: resolver,
		mutation: &mutationResolver{Resolver: resolver},
		query:    &queryResolver{Resolver: resolver},
		apiary:   &apiaryResolver{Resolver: resolver},
		hive:     &hiveResolver{Resolver: resolver},
		box:      &boxResolver{Resolver: resolver},
		frame:    &frameResolver{Resolver: resolver},
		family:   &familyResolver{Resolver: resolver},
	}

	if !seed {
		return fx
	}

	fx.apiaryID = createTestApiary(t, db, userID)
	fx.hiveID = createTestHive(t, db, userID, fx.apiaryID)
	fx.familyID = createTestQueen(t, db, userID, fx.hiveID)
	fx.boxID = createTestBox(t, db, userID, fx.hiveID)
	fx.leftSideID = createTestFrameSide(t, db, userID)
	fx.rightSideID = createTestFrameSide(t, db, userID)
	fx.frameID = createTestFrameWithSides(t, db, userID, fx.boxID, 1, fx.leftSideID, fx.rightSideID)

	_, err := db.Exec(
		"INSERT INTO family_moves (user_id, family_id, from_hive_id, to_hive_id, move_type) VALUES (?, ?, NULL, ?, 'ASSIGNED')",
		userID, fx.familyID, fx.hiveID,
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO treatments (type, hive_id, user_id, family_id) VALUES ('oxalic_acid', ?, ?, ?)",
		fx.hiveID, userID, fx.familyID,
	)
	require.NoError(t, err)

	inspectionID, err := (&model.Inspection{Db: db, UserID: userID}).Create("{}", fx.hiveID)
	require.NoError(t, err)
	fx.inspectionID = *inspectionID

	_, err = (&model.HivePlacement{Db: db, UserID: userID}).Update(strconv.Itoa(fx.apiaryID), strconv.Itoa(fx.hiveID), 10.5, 20.25, 90)
	require.NoError(t, err)

	_, err = (&model.ApiaryObstacle{Db: db, UserID: userID}).Create(strconv.Itoa(fx.apiaryID), model.ObstacleTypeCircle.String(), 1.5, 2.5, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	_, err = (&model.HiveLog{Db: db, UserID: userID}).Create(model.HiveLogInput{HiveID: strconv.Itoa(fx.hiveID), Action: "TEST", Title: "log"})
	require.NoError(t, err)

	return fx
}

func TestSchemaResolversMutations(t *testing.T) {
	t.Parallel()

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()

		t.Run("Add", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)

			// ACT
			created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "Resolver Apiary"})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, created)
		})

		t.Run("Update", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "Before"})
			require.NoError(t, err)

			// ACT
			updated, err := fx.mutation.UpdateApiary(fx.ctx, strconv.Itoa(created.ID), model.ApiaryInput{Name: "After"})

			// ASSERT
			require.NoError(t, err)
			assert.Equal(t, "After", *updated.Name)
		})

		t.Run("Deactivate", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "ToDeactivate"})
			require.NoError(t, err)

			// ACT
			ok, err := fx.mutation.DeactivateApiary(fx.ctx, strconv.Itoa(created.ID))

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, ok)
			assert.True(t, *ok)
		})
	})

	t.Run("HiveLog", func(t *testing.T) {
		t.Parallel()

		t.Run("AddUpdateDelete", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			details := "initial details"
			source := "integration-test"
			dedupeKey := "schema-resolver-hive-log-1"
			relatedHive := &model.HiveLogRelatedHiveInput{ID: strconv.Itoa(fx.hiveID)}

			// ACT
			created, createErr := fx.mutation.AddHiveLog(fx.ctx, model.HiveLogInput{
				HiveID:       strconv.Itoa(fx.hiveID),
				Action:       "INSPECTION",
				Title:        "Initial title",
				Details:      &details,
				Source:       &source,
				DedupeKey:    &dedupeKey,
				RelatedHives: []*model.HiveLogRelatedHiveInput{relatedHive},
			})
			updatedTitle := "Updated title"
			updatedDetails := "updated details"
			updated, updateErr := fx.mutation.UpdateHiveLog(fx.ctx, created.ID, model.HiveLogUpdateInput{
				Title:   &updatedTitle,
				Details: &updatedDetails,
			})
			deleted, deleteErr := fx.mutation.DeleteHiveLog(fx.ctx, created.ID)
			items, listErr := fx.query.HiveLogs(fx.ctx, strconv.Itoa(fx.hiveID), nil)

			// ASSERT
			require.NoError(t, createErr)
			require.NotNil(t, created)
			assert.Equal(t, "Initial title", created.Title)
			assert.Equal(t, "INSPECTION", created.Action)
			require.NoError(t, updateErr)
			require.NotNil(t, updated)
			assert.Equal(t, updatedTitle, updated.Title)
			assert.Equal(t, updatedDetails, *updated.Details)
			require.NoError(t, deleteErr)
			assert.True(t, deleted)
			require.NoError(t, listErr)
			assert.False(t, hasHiveLogID(items, created.ID))
		})

		t.Run("AddFailsForUnknownHive", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)

			// ACT
			_, err := fx.mutation.AddHiveLog(fx.ctx, model.HiveLogInput{
				HiveID: "99999999",
				Action: "INSPECTION",
				Title:  "Should fail",
			})

			// ASSERT
			require.Error(t, err)
			assert.ErrorContains(t, err, "hive not found")
		})
	})

	t.Run("QueenWarehouse", func(t *testing.T) {
		t.Parallel()

		t.Run("MoveToWarehouseAndAssignBack", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			hiveID := strconv.Itoa(fx.hiveID)
			familyID := strconv.Itoa(fx.familyID)

			// ACT
			moved, moveErr := fx.mutation.MoveQueenToWarehouse(fx.ctx, hiveID, familyID)
			warehouseQueens, warehouseErr := fx.query.WarehouseQueens(fx.ctx)
			assigned, assignErr := fx.mutation.AssignQueenFromWarehouse(fx.ctx, hiveID, familyID)

			// ASSERT
			require.NoError(t, moveErr)
			require.NotNil(t, moved)
			assert.Nil(t, moved.HiveID)
			require.NoError(t, warehouseErr)
			assert.True(t, hasFamilyID(warehouseQueens, familyID))
			require.NoError(t, assignErr)
			require.NotNil(t, assigned)
			require.NotNil(t, assigned.HiveID)
			assert.Equal(t, fx.hiveID, *assigned.HiveID)
		})
	})

	t.Run("Device", func(t *testing.T) {
		t.Parallel()

		t.Run("AddUpdateDeactivate", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			hiveID := strconv.Itoa(fx.hiveID)
			boxID := strconv.Itoa(fx.boxID)
			token := " token-1 "

			// ACT
			created, createErr := fx.mutation.AddDevice(fx.ctx, model.DeviceInput{
				Name:     " Hive sensor ",
				Type:     model.DeviceTypeIotSensor,
				APIToken: &token,
				HiveID:   &hiveID,
				BoxID:    &boxID,
			})
			newName := "Updated sensor"
			newType := model.DeviceTypeVideoCamera
			empty := ""
			updated, updateErr := fx.mutation.UpdateDevice(fx.ctx, created.ID, model.DeviceUpdateInput{
				Name:     &newName,
				Type:     &newType,
				APIToken: &empty,
				HiveID:   &empty,
			})
			deactivated, deactivateErr := fx.mutation.DeactivateDevice(fx.ctx, created.ID)
			items, listErr := fx.query.Devices(fx.ctx)

			// ASSERT
			require.NoError(t, createErr)
			require.NotNil(t, created)
			require.NotNil(t, created.BoxID)
			require.NotNil(t, created.HiveID)
			assert.Equal(t, fx.boxID, *created.BoxID)
			assert.Equal(t, fx.hiveID, *created.HiveID)
			require.NoError(t, updateErr)
			require.NotNil(t, updated)
			assert.Equal(t, "Updated sensor", updated.Name)
			assert.Equal(t, model.DeviceTypeVideoCamera, updated.Type)
			assert.Nil(t, updated.APIToken)
			assert.Nil(t, updated.HiveID)
			assert.Nil(t, updated.BoxID)
			require.NoError(t, deactivateErr)
			require.NotNil(t, deactivated)
			assert.True(t, *deactivated)
			require.NoError(t, listErr)
			assert.False(t, hasDeviceID(items, created.ID))
		})

		t.Run("AddFailsWhenBoxHiveMismatch", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			otherHiveID := createTestHive(t, fx.resolver.Db, fx.userID, fx.apiaryID)
			otherHiveIDString := strconv.Itoa(otherHiveID)
			boxID := strconv.Itoa(fx.boxID)

			// ACT
			_, err := fx.mutation.AddDevice(fx.ctx, model.DeviceInput{
				Name:   "Mismatch device",
				Type:   model.DeviceTypeIotSensor,
				HiveID: &otherHiveIDString,
				BoxID:  &boxID,
			})

			// ASSERT
			require.Error(t, err)
			assert.ErrorContains(t, err, "selected box does not belong to selected hive")
		})
	})

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

func TestSchemaResolversQueries(t *testing.T) {
	t.Parallel()

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()

		t.Run("Apiaries", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.Apiaries(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("Apiary", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.Apiary(fx.ctx, strconv.Itoa(fx.apiaryID))

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("ApiaryObstacles", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.ApiaryObstacles(fx.ctx, strconv.Itoa(fx.apiaryID))

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})

	t.Run("Hive", func(t *testing.T) {
		t.Parallel()

		t.Run("Hive", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.Hive(fx.ctx, strconv.Itoa(fx.hiveID))

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("HivePlacements", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.HivePlacements(fx.ctx, strconv.Itoa(fx.apiaryID))

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("HiveLogs", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.HiveLogs(fx.ctx, strconv.Itoa(fx.hiveID), nil)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})
	})

	t.Run("Frame", func(t *testing.T) {
		t.Parallel()

		t.Run("HiveFrame", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			_, invalidErr := fx.query.HiveFrame(fx.ctx, "not-a-number")
			item, err := fx.query.HiveFrame(fx.ctx, strconv.Itoa(fx.frameID))

			// ASSERT
			require.Error(t, invalidErr)
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("HiveFrameSide", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			_, invalidErr := fx.query.HiveFrameSide(fx.ctx, "invalid")
			item, err := fx.query.HiveFrameSide(fx.ctx, strconv.Itoa(fx.leftSideID))

			// ASSERT
			require.Error(t, invalidErr)
			require.NoError(t, err)
			require.NotNil(t, item)
		})
	})

	t.Run("Inspection", func(t *testing.T) {
		t.Parallel()

		t.Run("Inspection", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.Inspection(fx.ctx, fx.inspectionID)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("Inspections", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.Inspections(fx.ctx, strconv.Itoa(fx.hiveID), nil)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})

	t.Run("Warehouse", func(t *testing.T) {
		t.Parallel()

		t.Run("Devices", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.Devices(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("WarehouseModules", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.WarehouseModules(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("WarehouseInventory", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.WarehouseInventory(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("WarehouseSettings", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.WarehouseSettings(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("WarehouseModuleStats", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.WarehouseModuleStats(fx.ctx, model.WarehouseModuleTypeRoof)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("WarehouseInventoryStats", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.WarehouseInventoryStats(fx.ctx, "BOX:ROOF")

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("WarehouseQueens", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.WarehouseQueens(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})
	})

	t.Run("BoxSystem", func(t *testing.T) {
		t.Parallel()

		t.Run("BoxSystems", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.BoxSystems(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("FrameSpecs", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.FrameSpecs(fx.ctx, nil)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("BoxSpecs", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			systems, err := fx.query.BoxSystems(fx.ctx)
			require.NoError(t, err)
			require.NotEmpty(t, systems)

			// ACT
			items, err := fx.query.BoxSpecs(fx.ctx, systems[0].ID)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("BoxSystemFrameSettings", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.BoxSystemFrameSettings(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})
	})

	t.Run("Naming", func(t *testing.T) {
		t.Parallel()

		t.Run("RandomHiveName", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			name, err := fx.query.RandomHiveName(fx.ctx, nil)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, name)
		})

		t.Run("RandomHiveNameFallbackToBee", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			resolver := &queryResolver{Resolver: &Resolver{femaleNamesMap: map[string][]string{}}}

			// ACT
			name, err := resolver.RandomHiveName(context.Background(), nil)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, name)
			assert.Equal(t, "Bee", *name)
		})
	})
}

func TestSchemaResolversFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()

		t.Run("Hives", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.apiary.Hives(fx.ctx, &model.Apiary{ID: fx.apiaryID}, nil, nil)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})

	t.Run("Box", func(t *testing.T) {
		t.Parallel()

		t.Run("Frames", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.box.Frames(fx.ctx, &model.Box{ID: ptr(strconv.Itoa(fx.boxID))})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})

	t.Run("Hive", func(t *testing.T) {
		t.Parallel()

		t.Run("Boxes", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.hive.Boxes(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("Family", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.hive.Family(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("Families", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.hive.Families(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("BoxCount", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			count, err := fx.hive.BoxCount(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			assert.GreaterOrEqual(t, count, 1)
		})

		t.Run("InspectionCount", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			count, err := fx.hive.InspectionCount(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			assert.GreaterOrEqual(t, count, 1)
		})

		t.Run("LastInspection", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.hive.LastInspection(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("HiveType", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)

			// ACT
			defaultType, defaultErr := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: ""})
			invalidType, invalidErr := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: "bad"})
			validType, validErr := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: "HORIZONTAL"})

			// ASSERT
			require.NoError(t, defaultErr)
			require.NoError(t, invalidErr)
			require.NoError(t, validErr)
			assert.Equal(t, model.HiveTypeVertical, defaultType)
			assert.Equal(t, model.HiveTypeVertical, invalidType)
			assert.Equal(t, model.HiveTypeHorizontal, validType)
		})

		t.Run("IsNew", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			invalidDate := "not-a-date"
			freshDate := time.Now().Format(time.DateTime)

			// ACT
			isNewWithoutDate, noDateErr := fx.hive.IsNew(fx.ctx, &model.Hive{Added: nil})
			_, invalidErr := fx.hive.IsNew(fx.ctx, &model.Hive{Added: &invalidDate})
			isNewWithFreshDate, freshDateErr := fx.hive.IsNew(fx.ctx, &model.Hive{Added: &freshDate})

			// ASSERT
			require.NoError(t, noDateErr)
			require.Error(t, invalidErr)
			require.NoError(t, freshDateErr)
			assert.False(t, isNewWithoutDate)
			assert.True(t, isNewWithFreshDate)
		})
	})

	t.Run("Frame", func(t *testing.T) {
		t.Parallel()

		t.Run("LeftSide", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.frame.LeftSide(fx.ctx, &model.Frame{LeftID: &fx.leftSideID})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("RightSide", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.frame.RightSide(fx.ctx, &model.Frame{RightID: &fx.rightSideID})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})
	})

	t.Run("Family", func(t *testing.T) {
		t.Parallel()

		t.Run("LastTreatment", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.family.LastTreatment(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("Treatments", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.family.Treatments(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("LastHive", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.family.LastHive(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})
	})

	t.Run("ResolverFactory", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		fx := newSchemaResolverFixture(t, false)

		// ACT
		apiaryResolver := fx.resolver.Apiary()
		apiaryObstacleResolver := fx.resolver.ApiaryObstacle()
		boxResolver := fx.resolver.Box()
		familyResolver := fx.resolver.Family()
		frameResolver := fx.resolver.Frame()
		hiveResolver := fx.resolver.Hive()
		mutationResolver := fx.resolver.Mutation()
		queryResolver := fx.resolver.Query()

		// ASSERT
		assert.NotNil(t, apiaryResolver)
		assert.NotNil(t, apiaryObstacleResolver)
		assert.NotNil(t, boxResolver)
		assert.NotNil(t, familyResolver)
		assert.NotNil(t, frameResolver)
		assert.NotNil(t, hiveResolver)
		assert.NotNil(t, mutationResolver)
		assert.NotNil(t, queryResolver)
	})
}

func ptr[T any](v T) *T {
	return &v
}

func hasHiveLogID(logs []*model.HiveLog, id string) bool {
	for _, log := range logs {
		if log != nil && log.ID == id {
			return true
		}
	}
	return false
}

func hasFamilyID(families []*model.Family, id string) bool {
	for _, family := range families {
		if family != nil && family.ID == id {
			return true
		}
	}
	return false
}

func hasDeviceID(devices []*model.Device, id string) bool {
	for _, device := range devices {
		if device != nil && device.ID == id {
			return true
		}
	}
	return false
}

func hasHivePlacementID(placements []*model.HivePlacement, id string) bool {
	for _, placement := range placements {
		if placement != nil && placement.ID == id {
			return true
		}
	}
	return false
}
