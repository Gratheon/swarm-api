package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestDataLoader_BatchesHiveQueries(t *testing.T) {
	t.Parallel()

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

	if len(hives1) != 2 {
		t.Errorf("Expected 2 hives for apiary1, got %d", len(hives1))
	}

	if len(hives2) != 2 {
		t.Errorf("Expected 2 hives for apiary1 (second result), got %d", len(hives2))
	}

	foundHive1 := false
	foundHive2 := false
	for _, hive := range hives1 {
		hiveID := hive.ID
		if hiveID == fmt.Sprintf("%d", hive1ID) || hiveID == fmt.Sprintf("%d", hive2ID) {
			if hiveID == fmt.Sprintf("%d", hive1ID) {
				foundHive1 = true
			}
			if hiveID == fmt.Sprintf("%d", hive2ID) {
				foundHive2 = true
			}
		}
	}

	if !foundHive1 || !foundHive2 {
		t.Errorf("Expected to find both hive1 and hive2 in results")
	}
}

func TestDataLoader_BatchesBoxQueries(t *testing.T) {
	t.Parallel()

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

	box1ID := createTestBox(t, db, userID, hive1ID)
	box2ID := createTestBox(t, db, userID, hive1ID)
	box3ID := createTestBox(t, db, userID, hive2ID)

	ctx := context.WithValue(context.Background(), "userID", userID)

	loader := NewBoxLoader(db)

	resultChan1 := make(chan []*model.Box, 1)
	resultChan2 := make(chan []*model.Box, 1)

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

	if len(boxes1) != 2 {
		t.Errorf("Expected 2 boxes for hive1, got %d", len(boxes1))
	}

	if len(boxes2) != 1 {
		t.Errorf("Expected 1 box for hive2, got %d", len(boxes2))
	}

	_, _, _ = box1ID, box2ID, box3ID
}

func TestDataLoader_BatchesFamilyQueries(t *testing.T) {
	t.Parallel()

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

	queen1ID := createTestQueen(t, db, userID, hive1ID)
	queen2ID := createTestQueen(t, db, userID, hive2ID)

	ctx := context.WithValue(context.Background(), "userID", userID)

	loader := NewFamilyLoader(db)

	resultChan1 := make(chan *model.Family, 1)
	resultChan2 := make(chan *model.Family, 1)
	resultChan3 := make(chan *model.Family, 1)

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

	if family1 == nil {
		t.Errorf("Expected family for hive1, got nil")
	}

	if family2 == nil {
		t.Errorf("Expected family for hive2, got nil")
	}

	if family3 != nil {
		t.Errorf("Expected no family for hive3, got %v", family3)
	}

	_, _ = queen1ID, queen2ID
}

func TestDataLoader_BatchesFrameQueries(t *testing.T) {
	t.Parallel()

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

	frame1IDs := createTestFrames(t, db, userID, box1ID, 3)
	frame2IDs := createTestFrames(t, db, userID, box2ID, 2)

	ctx := context.WithValue(context.Background(), "userID", userID)

	loader := NewFrameLoader(db)

	resultChan1 := make(chan []*model.Frame, 1)
	resultChan2 := make(chan []*model.Frame, 1)

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

	if len(frames1) != 3 {
		t.Errorf("Expected 3 frames for box1, got %d", len(frames1))
	}

	if len(frames2) != 2 {
		t.Errorf("Expected 2 frames for box2, got %d", len(frames2))
	}

	_, _ = frame1IDs, frame2IDs
}

func TestDataLoader_IntegrationWithResolvers(t *testing.T) {
	t.Parallel()

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

	hives, err := apiaryResolver.Hives(ctx, apiary)
	if err != nil {
		t.Fatalf("Failed to resolve hives: %v", err)
	}

	if len(hives) != 2 {
		t.Errorf("Expected 2 hives, got %d", len(hives))
	}

	boxes1, err := hiveResolver.Boxes(ctx, hives[0])
	if err != nil {
		t.Fatalf("Failed to resolve boxes for hive1: %v", err)
	}

	if len(boxes1) == 0 {
		t.Errorf("Expected at least 1 box for hive1")
	}

	family1, err := hiveResolver.Family(ctx, hives[0])
	if err != nil {
		t.Fatalf("Failed to resolve family for hive1: %v", err)
	}

	if family1 == nil {
		t.Errorf("Expected family for hive1, got nil")
	}

	if len(boxes1) > 0 {
		frames, err := boxResolver.Frames(ctx, boxes1[0])
		if err != nil {
			t.Fatalf("Failed to resolve frames for box: %v", err)
		}

		if len(frames) == 0 {
			t.Errorf("Expected frames for box")
		}
	}
}

func TestDataLoader_FrameSidesLoading(t *testing.T) {
	t.Parallel()

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

	frame1ID := createTestFrameWithSides(t, db, userID, boxID, 0, leftSideID1, rightSideID1)
	frame2ID := createTestFrameWithSides(t, db, userID, boxID, 1, leftSideID2, rightSideID2)

	ctx := context.WithValue(context.Background(), "userID", userID)
	ctx = context.WithValue(ctx, LoadersKey, &Loaders{
		FrameSideLoader: NewFrameSideLoader(db),
	})

	loader := NewFrameSideLoader(db)

	resultChan1 := make(chan *model.FrameSide, 1)
	resultChan2 := make(chan *model.FrameSide, 1)

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

	if leftSide1 == nil {
		t.Error("Expected left side to be loaded, got nil")
	}

	if rightSide1 == nil {
		t.Error("Expected right side to be loaded, got nil")
	}

	if leftSide1 != nil && leftSide1.ID != nil {
		if *leftSide1.ID != fmt.Sprintf("%d", leftSideID1) {
			t.Errorf("Expected left side ID %d, got %s", leftSideID1, *leftSide1.ID)
		}
	}

	if rightSide1 != nil && rightSide1.ID != nil {
		if *rightSide1.ID != fmt.Sprintf("%d", rightSideID1) {
			t.Errorf("Expected right side ID %d, got %s", rightSideID1, *rightSide1.ID)
		}
	}

	_, _, _, _ = frame1ID, frame2ID, leftSideID2, rightSideID2
}

func TestDataLoader_FrameSidesWithResolvers(t *testing.T) {
	t.Parallel()

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

	leftSide, err := frameResolver.LeftSide(ctx, frame)
	if err != nil {
		t.Fatalf("Failed to resolve left side: %v", err)
	}

	if leftSide == nil {
		t.Fatal("Expected left side to be loaded, got nil")
	}

	if leftSide.ID == nil || *leftSide.ID != fmt.Sprintf("%d", leftSideID) {
		t.Errorf("Expected left side ID %d, got %v", leftSideID, leftSide.ID)
	}

	rightSide, err := frameResolver.RightSide(ctx, frame)
	if err != nil {
		t.Fatalf("Failed to resolve right side: %v", err)
	}

	if rightSide == nil {
		t.Fatal("Expected right side to be loaded, got nil")
	}

	if rightSide.ID == nil || *rightSide.ID != fmt.Sprintf("%d", rightSideID) {
		t.Errorf("Expected right side ID %d, got %v", rightSideID, rightSide.ID)
	}
}

func createTestFrameSide(t *testing.T, db *sqlx.DB, userID string) int {
	result := db.MustExec(
		"INSERT INTO frames_sides (user_id) VALUES (?)",
		userID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestFrameWithSides(t *testing.T, db *sqlx.DB, userID string, boxID int, position int, leftSideID int, rightSideID int) int {
	result := db.MustExec(
		"INSERT INTO frames (user_id, box_id, position, type, active, left_id, right_id) VALUES (?, ?, ?, 'EMPTY_COMB', 1, ?, ?)",
		userID, boxID, position, leftSideID, rightSideID,
	)
	id, _ := result.LastInsertId()
	return int(id)
}
