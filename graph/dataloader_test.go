package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	_ "github.com/go-sql-driver/mysql"
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
