package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestDataLoader(t *testing.T) {
	t.Parallel()

	t.Run("hive loader batches concurrent queries for same apiary", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999100"
		defer cleanupTestData(t, db, userID)

		apiary1ID := createTestApiary(t, db, userID)
		apiary2ID := createTestApiary(t, db, userID)

		hive1ID := createTestHive(t, db, userID, apiary1ID)
		hive2ID := createTestHive(t, db, userID, apiary1ID)
		_ = createTestHive(t, db, userID, apiary2ID)

		ctx := context.WithValue(context.Background(), "userID", userID)
		loader := NewHiveLoader(db)

		resultChan1 := make(chan []*model.Hive, 1)
		resultChan2 := make(chan []*model.Hive, 1)

		// ACT
		go func() {
			hives, err := loader.Load(ctx, apiary1ID, userID)
			if err != nil {
				t.Errorf("Failed to load hives for apiary1: %v", err)
			}
			resultChan1 <- hives
		}()

		go func() {
			hives, err := loader.Load(ctx, apiary1ID, userID)
			if err != nil {
				t.Errorf("Failed to load hives for apiary1 (second call): %v", err)
			}
			resultChan2 <- hives
		}()

		hives1 := <-resultChan1
		hives2 := <-resultChan2

		// ASSERT
		assert.Len(t, hives1, 2, "Expected 2 hives for apiary1")
		assert.Len(t, hives2, 2, "Expected 2 hives for apiary1 (second result)")

		foundHive1 := false
		foundHive2 := false
		for _, hive := range hives1 {
			hiveID := hive.ID
			if hiveID == fmt.Sprintf("%d", hive1ID) {
				foundHive1 = true
			}
			if hiveID == fmt.Sprintf("%d", hive2ID) {
				foundHive2 = true
			}
		}

		assert.True(t, foundHive1 && foundHive2, "Expected to find both hive1 and hive2 in results")
	})

	t.Run("box loader batches concurrent queries for different hives", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999101"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		hive1ID := createTestHive(t, db, userID, apiaryID)
		hive2ID := createTestHive(t, db, userID, apiaryID)

		createTestBox(t, db, userID, hive1ID)
		createTestBox(t, db, userID, hive1ID)
		createTestBox(t, db, userID, hive2ID)

		ctx := context.WithValue(context.Background(), "userID", userID)
		loader := NewBoxLoader(db)

		resultChan1 := make(chan []*model.Box, 1)
		resultChan2 := make(chan []*model.Box, 1)

		// ACT
		go func() {
			boxes, err := loader.Load(ctx, fmt.Sprintf("%d", hive1ID), userID)
			if err != nil {
				t.Errorf("Failed to load boxes for hive1: %v", err)
			}
			resultChan1 <- boxes
		}()

		go func() {
			boxes, err := loader.Load(ctx, fmt.Sprintf("%d", hive2ID), userID)
			if err != nil {
				t.Errorf("Failed to load boxes for hive2: %v", err)
			}
			resultChan2 <- boxes
		}()

		boxes1 := <-resultChan1
		boxes2 := <-resultChan2

		// ASSERT
		assert.Len(t, boxes1, 2, "Expected 2 boxes for hive1")
		assert.Len(t, boxes2, 1, "Expected 1 box for hive2")
	})

	t.Run("family loader batches queries and handles missing families", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999102"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		hive1ID := createTestHive(t, db, userID, apiaryID)
		hive2ID := createTestHive(t, db, userID, apiaryID)
		hive3ID := createTestHive(t, db, userID, apiaryID)

		createTestQueen(t, db, userID, hive1ID)
		createTestQueen(t, db, userID, hive2ID)

		ctx := context.WithValue(context.Background(), "userID", userID)
		loader := NewFamilyLoader(db)

		resultChan1 := make(chan *model.Family, 1)
		resultChan2 := make(chan *model.Family, 1)
		resultChan3 := make(chan *model.Family, 1)

		// ACT
		go func() {
			family, err := loader.Load(ctx, fmt.Sprintf("%d", hive1ID), userID)
			if err != nil {
				t.Errorf("Failed to load family for hive1: %v", err)
			}
			resultChan1 <- family
		}()

		go func() {
			family, err := loader.Load(ctx, fmt.Sprintf("%d", hive2ID), userID)
			if err != nil {
				t.Errorf("Failed to load family for hive2: %v", err)
			}
			resultChan2 <- family
		}()

		go func() {
			family, err := loader.Load(ctx, fmt.Sprintf("%d", hive3ID), userID)
			if err != nil {
				t.Errorf("Failed to load family for hive3: %v", err)
			}
			resultChan3 <- family
		}()

		family1 := <-resultChan1
		family2 := <-resultChan2
		family3 := <-resultChan3

		// ASSERT
		assert.NotNil(t, family1, "Expected family for hive1")
		assert.NotNil(t, family2, "Expected family for hive2")
		assert.Nil(t, family3, "Expected no family for hive3")
	})

	t.Run("frame loader batches concurrent queries for different boxes", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999103"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		hiveID := createTestHive(t, db, userID, apiaryID)
		box1ID := createTestBox(t, db, userID, hiveID)
		box2ID := createTestBox(t, db, userID, hiveID)

		createTestFrames(t, db, userID, box1ID, 3)
		createTestFrames(t, db, userID, box2ID, 2)

		ctx := context.WithValue(context.Background(), "userID", userID)
		loader := NewFrameLoader(db)

		resultChan1 := make(chan []*model.Frame, 1)
		resultChan2 := make(chan []*model.Frame, 1)

		// ACT
		go func() {
			frames, err := loader.Load(ctx, fmt.Sprintf("%d", box1ID), userID)
			if err != nil {
				t.Errorf("Failed to load frames for box1: %v", err)
			}
			resultChan1 <- frames
		}()

		go func() {
			frames, err := loader.Load(ctx, fmt.Sprintf("%d", box2ID), userID)
			if err != nil {
				t.Errorf("Failed to load frames for box2: %v", err)
			}
			resultChan2 <- frames
		}()

		frames1 := <-resultChan1
		frames2 := <-resultChan2

		// ASSERT
		assert.Len(t, frames1, 3, "Expected 3 frames for box1")
		assert.Len(t, frames2, 2, "Expected 2 frames for box2")
	})

	t.Run("frame side loader batches concurrent queries", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999105"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		hiveID := createTestHive(t, db, userID, apiaryID)
		boxID := createTestBox(t, db, userID, hiveID)

		leftSideID1 := createTestFrameSide(t, db, userID)
		rightSideID1 := createTestFrameSide(t, db, userID)
		leftSideID2 := createTestFrameSide(t, db, userID)
		rightSideID2 := createTestFrameSide(t, db, userID)

		createTestFrameWithSides(t, db, userID, boxID, 0, leftSideID1, rightSideID1)
		createTestFrameWithSides(t, db, userID, boxID, 1, leftSideID2, rightSideID2)

		ctx := context.WithValue(context.Background(), "userID", userID)
		loader := NewFrameSideLoader(db)

		resultChan1 := make(chan *model.FrameSide, 1)
		resultChan2 := make(chan *model.FrameSide, 1)

		// ACT
		go func() {
			leftSide, err := loader.Load(ctx, &leftSideID1, userID)
			if err != nil {
				t.Errorf("Failed to load left side for frame1: %v", err)
			}
			resultChan1 <- leftSide
		}()

		go func() {
			rightSide, err := loader.Load(ctx, &rightSideID1, userID)
			if err != nil {
				t.Errorf("Failed to load right side for frame1: %v", err)
			}
			resultChan2 <- rightSide
		}()

		leftSide1 := <-resultChan1
		rightSide1 := <-resultChan2

		// ASSERT
		assert.NotNil(t, leftSide1, "Expected left side to be loaded")
		assert.NotNil(t, rightSide1, "Expected right side to be loaded")

		if leftSide1 != nil && leftSide1.ID != nil {
			assert.Equal(t, fmt.Sprintf("%d", leftSideID1), *leftSide1.ID, "Left side ID should match")
		}

		if rightSide1 != nil && rightSide1.ID != nil {
			assert.Equal(t, fmt.Sprintf("%d", rightSideID1), *rightSide1.ID, "Right side ID should match")
		}
	})
}

func TestDataLoaderIntegrationWithResolvers(t *testing.T) {
	t.Parallel()

	t.Run("dataloaders work correctly with GraphQL resolver chain", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999104"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		hive1ID := createTestHive(t, db, userID, apiaryID)
		hive2ID := createTestHive(t, db, userID, apiaryID)

		box1ID := createTestBox(t, db, userID, hive1ID)
		box2ID := createTestBox(t, db, userID, hive2ID)

		createTestFrames(t, db, userID, box1ID, 2)
		createTestFrames(t, db, userID, box2ID, 3)

		createTestQueen(t, db, userID, hive1ID)
		createTestQueen(t, db, userID, hive2ID)

		ctx := context.WithValue(context.Background(), "userID", userID)
		ctx = context.WithValue(ctx, LoadersKey, &Loaders{
			HivesByApiaryLoader: NewHiveLoader(db),
			BoxesByHiveLoader:   NewBoxLoader(db),
			FamilyByHiveLoader:  NewFamilyLoader(db),
			FramesByBoxLoader:   NewFrameLoader(db),
		})

		resolver := &Resolver{Db: db}
		apiaryResolver := &apiaryResolver{Resolver: resolver}
		hiveResolver := &hiveResolver{Resolver: resolver}
		boxResolver := &boxResolver{Resolver: resolver}

		apiary := &model.Apiary{ID: apiaryID}

		// ACT
		hives, err := apiaryResolver.Hives(ctx, apiary)
		assert.NoError(t, err, "Failed to resolve hives")

		boxes1, err := hiveResolver.Boxes(ctx, hives[0])
		assert.NoError(t, err, "Failed to resolve boxes for hive1")

		family1, err := hiveResolver.Family(ctx, hives[0])
		assert.NoError(t, err, "Failed to resolve family for hive1")

		var frames []*model.Frame
		if len(boxes1) > 0 {
			frames, err = boxResolver.Frames(ctx, boxes1[0])
			assert.NoError(t, err, "Failed to resolve frames for box")
		}

		// ASSERT
		assert.Len(t, hives, 2, "Expected 2 hives")
		assert.Greater(t, len(boxes1), 0, "Expected at least 1 box for hive1")
		assert.NotNil(t, family1, "Expected family for hive1")
		assert.Greater(t, len(frames), 0, "Expected frames for box")
	})

	t.Run("frame side loader works with frame resolver", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		db := setupTestDB(t)
		if db == nil {
			return
		}
		defer db.Close()

		userID := "999106"
		defer cleanupTestData(t, db, userID)

		apiaryID := createTestApiary(t, db, userID)
		hiveID := createTestHive(t, db, userID, apiaryID)
		boxID := createTestBox(t, db, userID, hiveID)

		leftSideID := createTestFrameSide(t, db, userID)
		rightSideID := createTestFrameSide(t, db, userID)
		frameID := createTestFrameWithSides(t, db, userID, boxID, 0, leftSideID, rightSideID)

		ctx := context.WithValue(context.Background(), "userID", userID)
		ctx = context.WithValue(ctx, LoadersKey, &Loaders{
			FrameSideLoader: NewFrameSideLoader(db),
		})

		resolver := &Resolver{Db: db}
		frameResolver := &frameResolver{Resolver: resolver}

		frame := &model.Frame{
			Db:      db,
			ID:      frameID,
			UserID:  userID,
			LeftID:  &leftSideID,
			RightID: &rightSideID,
		}

		// ACT
		leftSide, err := frameResolver.LeftSide(ctx, frame)
		assert.NoError(t, err, "Failed to resolve left side")

		rightSide, err := frameResolver.RightSide(ctx, frame)
		assert.NoError(t, err, "Failed to resolve right side")

		// ASSERT
		assert.NotNil(t, leftSide, "Expected left side to be loaded")
		assert.NotNil(t, rightSide, "Expected right side to be loaded")

		if leftSide != nil && leftSide.ID != nil {
			assert.Equal(t, fmt.Sprintf("%d", leftSideID), *leftSide.ID, "Left side ID should match")
		}

		if rightSide != nil && rightSide.ID != nil {
			assert.Equal(t, fmt.Sprintf("%d", rightSideID), *rightSide.ID, "Right side ID should match")
		}
	})
}
