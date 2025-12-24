package graph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/jmoiron/sqlx"
)

type contextKey string

const LoadersKey = contextKey("dataloader")

type Loaders struct {
	HivesByApiaryLoader *HiveLoader
	BoxesByHiveLoader   *BoxLoader
	FamilyByHiveLoader  *FamilyLoader
	FramesByBoxLoader   *FrameLoader
	FrameSideLoader     *FrameSideLoader
}

func GetLoaders(ctx context.Context) *Loaders {
	val := ctx.Value(LoadersKey)
	if val == nil {
		return nil
	}
	return val.(*Loaders)
}

type HiveLoader struct {
	db    *sqlx.DB
	mu    sync.Mutex
	batch map[int]*hiveBatch
	wait  time.Duration
}

type hiveBatch struct {
	keys     []int
	channels []chan []*model.Hive
	userID   string
}

func NewHiveLoader(db *sqlx.DB) *HiveLoader {
	return &HiveLoader{
		db:    db,
		batch: make(map[int]*hiveBatch),
		wait:  1 * time.Millisecond,
	}
}

func (l *HiveLoader) Load(ctx context.Context, apiaryID int, userID string) ([]*model.Hive, error) {
	resultChan := make(chan []*model.Hive, 1)

	l.mu.Lock()
	if l.batch[apiaryID] == nil {
		l.batch[apiaryID] = &hiveBatch{
			keys:     []int{apiaryID},
			channels: []chan []*model.Hive{resultChan},
			userID:   userID,
		}
		go l.processBatch(apiaryID, l.wait)
	} else {
		l.batch[apiaryID].channels = append(l.batch[apiaryID].channels, resultChan)
	}
	l.mu.Unlock()

	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *HiveLoader) processBatch(apiaryID int, wait time.Duration) {
	time.Sleep(wait)

	l.mu.Lock()
	batch := l.batch[apiaryID]
	delete(l.batch, apiaryID)
	l.mu.Unlock()

	if batch == nil || len(batch.channels) == 0 {
		return
	}

	hives := []*model.Hive{}
	err := l.db.Select(&hives,
		`SELECT id, user_id, apiary_id, active, hive_number, notes, color, status, added, 
		        collapse_date, collapse_cause, parent_hive_id, split_date, merged_into_hive_id, merge_date, merge_type
		FROM hives 
		WHERE apiary_id=? AND user_id=? AND active=1 AND collapse_date IS NULL AND merged_into_hive_id IS NULL`,
		apiaryID, batch.userID)

	if err != nil {
		for _, ch := range batch.channels {
			close(ch)
		}
		return
	}

	for _, ch := range batch.channels {
		ch <- hives
	}
}

type BoxLoader struct {
	db    *sqlx.DB
	mu    sync.Mutex
	batch map[string]chan []*model.Box
	timer *time.Timer
	wait  time.Duration
}

func NewBoxLoader(db *sqlx.DB) *BoxLoader {
	return &BoxLoader{
		db:    db,
		batch: make(map[string]chan []*model.Box),
		wait:  1 * time.Millisecond,
	}
}

func (l *BoxLoader) Load(ctx context.Context, hiveID string, userID string) ([]*model.Box, error) {
	resultChan := make(chan []*model.Box, 1)

	l.mu.Lock()
	needsScheduling := len(l.batch) == 0
	l.batch[hiveID] = resultChan

	if needsScheduling {
		l.timer = time.AfterFunc(l.wait, func() {
			l.processBatch(userID)
		})
	}
	l.mu.Unlock()

	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *BoxLoader) processBatch(userID string) {
	l.mu.Lock()
	batch := l.batch
	l.batch = make(map[string]chan []*model.Box)
	l.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	hiveIDs := make([]string, 0, len(batch))
	for hiveID := range batch {
		hiveIDs = append(hiveIDs, hiveID)
	}

	query, args, err := sqlx.In(
		`SELECT * FROM boxes 
		WHERE active=1 AND hive_id IN (?) AND user_id=? 
		ORDER BY hive_id, position DESC`,
		hiveIDs, userID)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	query = l.db.Rebind(query)

	var allBoxes []*model.Box
	err = l.db.Select(&allBoxes, query, args...)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	boxesByHive := make(map[string][]*model.Box)
	for _, box := range allBoxes {
		hiveIDStr := fmt.Sprintf("%d", box.HiveId)
		boxesByHive[hiveIDStr] = append(boxesByHive[hiveIDStr], box)
	}

	for hiveID, ch := range batch {
		boxes := boxesByHive[hiveID]
		if boxes == nil {
			boxes = []*model.Box{}
		}
		ch <- boxes
	}
}

type FamilyLoader struct {
	db    *sqlx.DB
	mu    sync.Mutex
	batch map[string]chan *model.Family
	timer *time.Timer
	wait  time.Duration
}

func NewFamilyLoader(db *sqlx.DB) *FamilyLoader {
	return &FamilyLoader{
		db:    db,
		batch: make(map[string]chan *model.Family),
		wait:  1 * time.Millisecond,
	}
}

func (l *FamilyLoader) Load(ctx context.Context, hiveID string, userID string) (*model.Family, error) {
	resultChan := make(chan *model.Family, 1)

	l.mu.Lock()
	needsScheduling := len(l.batch) == 0
	l.batch[hiveID] = resultChan

	if needsScheduling {
		l.timer = time.AfterFunc(l.wait, func() {
			l.processBatch(userID)
		})
	}
	l.mu.Unlock()

	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *FamilyLoader) processBatch(userID string) {
	l.mu.Lock()
	batch := l.batch
	l.batch = make(map[string]chan *model.Family)
	l.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	hiveIDs := make([]string, 0, len(batch))
	for hiveID := range batch {
		hiveIDs = append(hiveIDs, hiveID)
	}

	query, args, err := sqlx.In(
		`SELECT * FROM families 
		WHERE hive_id IN (?) AND user_id=?`,
		hiveIDs, userID)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	query = l.db.Rebind(query)

	var allFamilies []*model.Family
	err = l.db.Select(&allFamilies, query, args...)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	familiesByHive := make(map[string]*model.Family)
	for _, family := range allFamilies {
		if family.HiveID != nil {
			hiveIDStr := fmt.Sprintf("%d", *family.HiveID)
			if _, exists := familiesByHive[hiveIDStr]; !exists {
				familiesByHive[hiveIDStr] = family
			}
		}
	}

	for hiveID, ch := range batch {
		family := familiesByHive[hiveID]
		ch <- family
	}
}

type FrameLoader struct {
	db    *sqlx.DB
	mu    sync.Mutex
	batch map[string]chan []*model.Frame
	timer *time.Timer
	wait  time.Duration
}

func NewFrameLoader(db *sqlx.DB) *FrameLoader {
	return &FrameLoader{
		db:    db,
		batch: make(map[string]chan []*model.Frame),
		wait:  1 * time.Millisecond,
	}
}

func (l *FrameLoader) Load(ctx context.Context, boxID string, userID string) ([]*model.Frame, error) {
	resultChan := make(chan []*model.Frame, 1)

	l.mu.Lock()
	needsScheduling := len(l.batch) == 0
	l.batch[boxID] = resultChan

	if needsScheduling {
		l.timer = time.AfterFunc(l.wait, func() {
			l.processBatch(userID)
		})
	}
	l.mu.Unlock()

	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *FrameLoader) processBatch(userID string) {
	l.mu.Lock()
	batch := l.batch
	l.batch = make(map[string]chan []*model.Frame)
	l.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	boxIDs := make([]string, 0, len(batch))
	for boxID := range batch {
		boxIDs = append(boxIDs, boxID)
	}

	query, args, err := sqlx.In(
		`SELECT * FROM frames 
		WHERE active=1 AND box_id IN (?) AND user_id=? 
		ORDER BY box_id, position`,
		boxIDs, userID)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	query = l.db.Rebind(query)

	var allFrames []*model.Frame
	err = l.db.Select(&allFrames, query, args...)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	framesByBox := make(map[string][]*model.Frame)
	for _, frame := range allFrames {
		boxIDStr := fmt.Sprintf("%d", frame.BoxId)
		framesByBox[boxIDStr] = append(framesByBox[boxIDStr], frame)
	}

	for boxID, ch := range batch {
		frames := framesByBox[boxID]
		if frames == nil {
			frames = []*model.Frame{}
		}
		ch <- frames
	}
}

type FrameSideLoader struct {
	db    *sqlx.DB
	mu    sync.Mutex
	batch map[int]chan *model.FrameSide
	timer *time.Timer
	wait  time.Duration
}

func NewFrameSideLoader(db *sqlx.DB) *FrameSideLoader {
	return &FrameSideLoader{
		db:    db,
		batch: make(map[int]chan *model.FrameSide),
		wait:  1 * time.Millisecond,
	}
}

func (l *FrameSideLoader) Load(ctx context.Context, frameSideID *int, userID string) (*model.FrameSide, error) {
	if frameSideID == nil {
		return nil, nil
	}

	resultChan := make(chan *model.FrameSide, 1)

	l.mu.Lock()
	needsScheduling := len(l.batch) == 0
	l.batch[*frameSideID] = resultChan

	if needsScheduling {
		l.timer = time.AfterFunc(l.wait, func() {
			l.processBatch(userID)
		})
	}
	l.mu.Unlock()

	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *FrameSideLoader) processBatch(userID string) {
	l.mu.Lock()
	batch := l.batch
	l.batch = make(map[int]chan *model.FrameSide)
	l.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	frameSideIDs := make([]int, 0, len(batch))
	for frameSideID := range batch {
		frameSideIDs = append(frameSideIDs, frameSideID)
	}

	query, args, err := sqlx.In(
		`SELECT * FROM frames_sides 
		WHERE id IN (?) AND user_id=?`,
		frameSideIDs, userID)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	query = l.db.Rebind(query)

	var allFrameSides []*model.FrameSide
	err = l.db.Select(&allFrameSides, query, args...)

	if err != nil {
		for _, ch := range batch {
			close(ch)
		}
		return
	}

	frameSidesMap := make(map[int]*model.FrameSide)
	for _, frameSide := range allFrameSides {
		if frameSide.ID != nil {
			var id int
			fmt.Sscanf(*frameSide.ID, "%d", &id)
			frameSidesMap[id] = frameSide
		}
	}

	for frameSideID, ch := range batch {
		frameSide := frameSidesMap[frameSideID]
		ch <- frameSide
	}
}
