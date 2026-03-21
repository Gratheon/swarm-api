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
