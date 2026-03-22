package graph

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Gratheon/log-lib-go"
	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/Gratheon/swarm-api/redisPubSub"
)

// AddHive is the resolver for the addHive field.
func (r *mutationResolver) AddHive(ctx context.Context, hive model.HiveInput) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	if hive.HiveType == nil {
		defaultHiveType := model.HiveTypeVertical
		hive.HiveType = &defaultHiveType
	}
	if hive.HiveType != nil && *hive.HiveType == model.HiveTypeHorizontal {
		horizontalType := model.BoxTypeLargeHorizontalSection
		hive.InitialBoxType = &horizontalType
	}
	if hive.HiveType != nil && *hive.HiveType == model.HiveTypeNucleus {
		deepType := model.BoxTypeDeep
		hive.InitialBoxType = &deepType
		hive.BoxCount = 1
	}

	if err := enforceHiveCreationLimit(ctx, hiveModel); err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	race := "unknown"
	var added string
	if hive.QueenYear != nil && *hive.QueenYear != "" {
		added = *hive.QueenYear
	} else {
		year := time.Now().Year()
		added = strconv.Itoa(year)
	}

	hiveResult, err := hiveModel.Create(hive)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	_, err = (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).CreateForHive(hiveResult.ID, hive.QueenName, &race, &added, hive.QueenColor)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
	}

	err = (&model.Box{
		Db:     r.Db,
		UserID: uid,
	}).CreateByHiveId(hiveResult.ID, hive.BoxCount, hive.Colors, func() model.BoxType {
		if hive.InitialBoxType != nil {
			return *hive.InitialBoxType
		}
		return model.BoxTypeDeep
	}())

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		_, deactivateErr := (&model.Hive{
			Db:     r.Resolver.Db,
			UserID: uid,
		}).Deactivate(hiveResult.ID)
		if deactivateErr != nil {
			logger.ErrorWithContext(ctx, deactivateErr.Error())
		}
		return nil, fmt.Errorf("failed to create hive sections: %w", err)
	}

	boxes, err := (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(hiveResult.ID)

	if err == nil {
		for _, box := range boxes {
			if box.Type == model.BoxTypeRoof {
				continue
			}
			err = (&model.Frame{
				Db:     r.Db,
				UserID: uid,
			}).CreateFramesForBox(box.ID, hive.FrameCount)
		}
	}

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return hiveResult, err
	}

	isNucleusHive := hive.HiveType != nil && *hive.HiveType == model.HiveTypeNucleus
	if !isNucleusHive {
		_, err = (&model.Box{
			Db:     r.Db,
			UserID: uid,
		}).CreateSingleBox(hiveResult.ID, hive.BoxCount, "#363636", model.BoxTypeRoof)
		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return hiveResult, fmt.Errorf("failed to create hive roof section: %w", err)
		}

		_, err = (&model.Box{
			Db:     r.Db,
			UserID: uid,
		}).CreateSingleBox(hiveResult.ID, -1, "#4a4a4a", model.BoxTypeBottom)
		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return hiveResult, fmt.Errorf("failed to create hive bottom section: %w", err)
		}
	}

	return hiveResult, err
}

// UpdateHive is the resolver for the updateHive field.
func (r *mutationResolver) UpdateHive(ctx context.Context, hive model.HiveUpdateInput) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyID, err := familyModel.Upsert(uid, hive)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
	}

	err = hiveModel.Update(hive.ID, hive.Notes, hive.HiveNumber, familyID)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
	}

	return hiveModel.Get(hive.ID)
}

// DeactivateHive is the resolver for the deactivateHive field.
func (r *mutationResolver) DeactivateHive(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}

// MarkHiveAsCollapsed is the resolver for the markHiveAsCollapsed field.
func (r *mutationResolver) MarkHiveAsCollapsed(ctx context.Context, id string, collapseDate string, collapseCause string) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	parsedCollapseDate, err := time.Parse(time.RFC3339, collapseDate)
	if err != nil {
		parsedCollapseDate, err = time.Parse("2006-01-02", collapseDate)
		if err != nil {
			logger.ErrorWithContext(ctx, "Invalid collapseDate format: "+err.Error())
			return nil, errors.New("Invalid collapseDate format, must be RFC3339 or YYYY-MM-DD")
		}
	}

	err = hiveModel.MarkAsCollapsed(id, parsedCollapseDate, collapseCause)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	updatedHive, err := hiveModel.Get(id)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return updatedHive, nil
}

// SplitHive is the resolver for the splitHive field.
func (r *mutationResolver) SplitHive(ctx context.Context, sourceHiveID string, queenName *string, queenAction string, frameIds []string) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)

	if len(frameIds) == 0 || len(frameIds) > 10 {
		return nil, errors.New("must select between 1 and 10 frames to split")
	}

	validQueenActions := map[string]bool{
		"new_queen":      true,
		"take_old_queen": true,
		"no_queen":       true,
	}
	if !validQueenActions[queenAction] {
		return nil, errors.New("invalid queenAction. Must be: new_queen, take_old_queen, or no_queen")
	}

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	if err := enforceHiveCreationLimit(ctx, hiveModel); err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	sourceHive, err := hiveModel.Get(sourceHiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}
	if sourceHive == nil {
		return nil, errors.New("source hive not found")
	}

	newHive, err := hiveModel.Split(sourceHiveID, sourceHive.ApiaryID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	if queenAction == "new_queen" {
		if queenName == nil || *queenName == "" {
			return nil, errors.New("queenName is required when queenAction is new_queen")
		}
		_, err = familyModel.CreateForHive(newHive.ID, queenName, nil, nil, nil)
		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return nil, err
		}
	} else if queenAction == "take_old_queen" {
		logger.Info("Attempting to take old queen from hive " + sourceHiveID + " for user " + uid)

		families, err := familyModel.ListByHive(sourceHiveID)
		if err != nil {
			logger.ErrorWithContext(ctx, "Error listing families for source hive "+sourceHiveID+": "+err.Error())
			return nil, err
		}

		logger.Info("Found " + strconv.Itoa(len(families)) + " queen(s) in source hive " + sourceHiveID)

		if len(families) == 0 {
			return nil, errors.New("source hive has no queen to take")
		}

		oldQueen := families[0]
		logger.Info("Moving queen " + oldQueen.ID + " from hive " + sourceHiveID + " to hive " + newHive.ID)

		err = familyModel.MoveBetweenHives(oldQueen.ID, sourceHiveID, newHive.ID)
		if err != nil {
			logger.ErrorWithContext(ctx, "Error moving queen to split hive: "+err.Error())
			return nil, err
		}
	}

	boxModel := &model.Box{
		Db:     r.Db,
		UserID: uid,
	}

	newBoxID, err := boxModel.CreateSingleBox(newHive.ID, 0, "#ffc848", model.BoxTypeDeep)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	_, err = boxModel.CreateSingleBox(newHive.ID, 1, "#363636", model.BoxTypeRoof)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	frameModel := &model.Frame{
		Db:     r.Db,
		UserID: uid,
	}

	err = frameModel.MoveFramesToBox(frameIds, newBoxID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	redisPubSub.PublishEvent(uid, "hive", newHive.ID, "split", newHive)

	return newHive, nil
}

// JoinHives is the resolver for the joinHives field.
func (r *mutationResolver) JoinHives(ctx context.Context, sourceHiveID string, targetHiveID string, mergeType string) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)

	validMergeTypes := map[string]bool{
		"both_queens":       true,
		"source_queen_kept": true,
		"target_queen_kept": true,
	}
	if !validMergeTypes[mergeType] {
		return nil, errors.New("invalid merge type")
	}

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	sourceHive, err := hiveModel.Get(sourceHiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}
	if sourceHive == nil {
		return nil, errors.New("source hive not found")
	}

	targetHive, err := hiveModel.Get(targetHiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}
	if targetHive == nil {
		return nil, errors.New("target hive not found")
	}

	boxModel := &model.Box{
		Db:     r.Db,
		UserID: uid,
	}

	sourceBoxes, err := boxModel.ListByHive(sourceHiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	boxesToKeepInSource, err := boxModel.GetBoxesByTypeForHive(sourceHiveID, []model.BoxType{model.BoxTypeBottom, model.BoxTypeGate, model.BoxTypeRoof})
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	keepInSourceMap := make(map[string]bool)
	for _, box := range boxesToKeepInSource {
		if box.ID != nil {
			keepInSourceMap[*box.ID] = true
		}
	}

	var boxIDsToMove []string
	for _, box := range sourceBoxes {
		if box.ID != nil && !keepInSourceMap[*box.ID] {
			boxIDsToMove = append(boxIDsToMove, *box.ID)
		}
	}

	if len(boxIDsToMove) > 0 {
		maxPos, err := boxModel.GetMaxPosition(targetHiveID)
		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return nil, err
		}

		err = boxModel.MoveBoxesToHive(boxIDsToMove, targetHiveID, maxPos+1)
		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return nil, err
		}
	}

	now := time.Now()
	err = hiveModel.MarkAsMerged(sourceHiveID, targetHiveID, now, mergeType)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	updatedTargetHive, err := hiveModel.Get(targetHiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	redisPubSub.PublishEvent(uid, "hive", targetHiveID, "join", updatedTargetHive)
	redisPubSub.PublishEvent(uid, "hive", sourceHiveID, "merged", sourceHive)

	return updatedTargetHive, nil
}
