//go:build integration
// +build integration

package graph

import (
	"context"
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
	if db == nil {
		return nil
	}
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

	t.Run("AddApiary", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, false)
		if fx == nil {
			return
		}
		created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "Resolver Apiary"})
		require.NoError(t, err)
		require.NotNil(t, created)
	})

	t.Run("UpdateApiary", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, false)
		if fx == nil {
			return
		}
		created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "Before"})
		require.NoError(t, err)
		updated, err := fx.mutation.UpdateApiary(fx.ctx, strconv.Itoa(created.ID), model.ApiaryInput{Name: "After"})
		require.NoError(t, err)
		assert.Equal(t, "After", *updated.Name)
	})

	t.Run("DeactivateApiary", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, false)
		if fx == nil {
			return
		}
		created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "ToDeactivate"})
		require.NoError(t, err)
		ok, err := fx.mutation.DeactivateApiary(fx.ctx, strconv.Itoa(created.ID))
		require.NoError(t, err)
		require.NotNil(t, ok)
		assert.True(t, *ok)
	})
}

func TestSchemaResolversQueries(t *testing.T) {
	t.Parallel()

	t.Run("Apiaries", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.Apiaries(fx.ctx)
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.query.Apiary(fx.ctx, strconv.Itoa(fx.apiaryID))
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Hive", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.query.Hive(fx.ctx, strconv.Itoa(fx.hiveID))
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("HiveFrame", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		_, err := fx.query.HiveFrame(fx.ctx, "not-a-number")
		require.Error(t, err)
		item, err := fx.query.HiveFrame(fx.ctx, strconv.Itoa(fx.frameID))
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("HiveFrameSide", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		_, err := fx.query.HiveFrameSide(fx.ctx, "invalid")
		require.Error(t, err)
		item, err := fx.query.HiveFrameSide(fx.ctx, strconv.Itoa(fx.leftSideID))
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Inspection", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.query.Inspection(fx.ctx, fx.inspectionID)
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Inspections", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.Inspections(fx.ctx, strconv.Itoa(fx.hiveID), nil)
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("HivePlacements", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.HivePlacements(fx.ctx, strconv.Itoa(fx.apiaryID))
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("ApiaryObstacles", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.ApiaryObstacles(fx.ctx, strconv.Itoa(fx.apiaryID))
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Devices", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.Devices(fx.ctx)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("WarehouseModules", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.WarehouseModules(fx.ctx)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("WarehouseInventory", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.WarehouseInventory(fx.ctx)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("WarehouseSettings", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.query.WarehouseSettings(fx.ctx)
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("WarehouseModuleStats", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.query.WarehouseModuleStats(fx.ctx, model.WarehouseModuleTypeRoof)
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("WarehouseInventoryStats", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.query.WarehouseInventoryStats(fx.ctx, "BOX:ROOF")
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("BoxSystems", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.BoxSystems(fx.ctx)
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("FrameSpecs", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.FrameSpecs(fx.ctx, nil)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("BoxSpecs", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		systems, err := fx.query.BoxSystems(fx.ctx)
		require.NoError(t, err)
		require.NotEmpty(t, systems)
		items, err := fx.query.BoxSpecs(fx.ctx, systems[0].ID)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("BoxSystemFrameSettings", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.BoxSystemFrameSettings(fx.ctx)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("WarehouseQueens", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.WarehouseQueens(fx.ctx)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("HiveLogs", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.query.HiveLogs(fx.ctx, strconv.Itoa(fx.hiveID), nil)
		require.NoError(t, err)
		assert.NotNil(t, items)
	})

	t.Run("RandomHiveName", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		name, err := fx.query.RandomHiveName(fx.ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, name)
	})

	t.Run("RandomHiveNameFallbackToBee", func(t *testing.T) {
		t.Parallel()
		resolver := &queryResolver{Resolver: &Resolver{femaleNamesMap: map[string][]string{}}}
		name, err := resolver.RandomHiveName(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, name)
		assert.Equal(t, "Bee", *name)
	})
}

func TestSchemaResolversFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Apiary.Hives", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.apiary.Hives(fx.ctx, &model.Apiary{ID: fx.apiaryID}, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Box.Frames", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.box.Frames(fx.ctx, &model.Box{ID: ptr(strconv.Itoa(fx.boxID))})
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Hive.Boxes", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.hive.Boxes(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Hive.Family", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.hive.Family(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Hive.Families", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.hive.Families(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Hive.BoxCount", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		count, err := fx.hive.BoxCount(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("Hive.InspectionCount", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		count, err := fx.hive.InspectionCount(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("Hive.LastInspection", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.hive.LastInspection(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Frame.LeftSide", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.frame.LeftSide(fx.ctx, &model.Frame{LeftID: &fx.leftSideID})
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Frame.RightSide", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.frame.RightSide(fx.ctx, &model.Frame{RightID: &fx.rightSideID})
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Family.LastTreatment", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.family.LastTreatment(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Family.Treatments", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		items, err := fx.family.Treatments(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})
		require.NoError(t, err)
		require.NotEmpty(t, items)
	})

	t.Run("Family.LastHive", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, true)
		if fx == nil {
			return
		}
		item, err := fx.family.LastHive(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})
		require.NoError(t, err)
		require.NotNil(t, item)
	})

	t.Run("Hive.HiveType", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, false)
		if fx == nil {
			return
		}
		defaultType, err := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: ""})
		require.NoError(t, err)
		assert.Equal(t, model.HiveTypeVertical, defaultType)
		invalidType, err := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: "bad"})
		require.NoError(t, err)
		assert.Equal(t, model.HiveTypeVertical, invalidType)
		validType, err := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: "HORIZONTAL"})
		require.NoError(t, err)
		assert.Equal(t, model.HiveTypeHorizontal, validType)
	})

	t.Run("Hive.IsNew", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, false)
		if fx == nil {
			return
		}
		isNew, err := fx.hive.IsNew(fx.ctx, &model.Hive{Added: nil})
		require.NoError(t, err)
		assert.False(t, isNew)
		invalidDate := "not-a-date"
		_, err = fx.hive.IsNew(fx.ctx, &model.Hive{Added: &invalidDate})
		require.Error(t, err)
		freshDate := time.Now().Format(time.DateTime)
		isNew, err = fx.hive.IsNew(fx.ctx, &model.Hive{Added: &freshDate})
		require.NoError(t, err)
		assert.True(t, isNew)
	})

	t.Run("ResolverFactories", func(t *testing.T) {
		t.Parallel()
		fx := newSchemaResolverFixture(t, false)
		if fx == nil {
			return
		}
		assert.NotNil(t, fx.resolver.Apiary())
		assert.NotNil(t, fx.resolver.ApiaryObstacle())
		assert.NotNil(t, fx.resolver.Box())
		assert.NotNil(t, fx.resolver.Family())
		assert.NotNil(t, fx.resolver.Frame())
		assert.NotNil(t, fx.resolver.Hive())
		assert.NotNil(t, fx.resolver.Mutation())
		assert.NotNil(t, fx.resolver.Query())
	})
}

func ptr[T any](v T) *T {
	return &v
}
