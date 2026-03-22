package graph

import (
	"context"
	"strconv"

	"github.com/Gratheon/log-lib-go"
	"github.com/Gratheon/swarm-api/graph/model"
)

// AddFrame is the resolver for the addFrame field.
func (r *mutationResolver) AddFrame(ctx context.Context, boxID string, typeArg string, position int) (*model.Frame, error) {
	uid := ctx.Value("userID").(string)
	frameType := model.FrameType(typeArg)

	frameModel := &model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	existingFrames, err := frameModel.ListByBox(&boxID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}
	occupiedPositions := map[int]bool{}
	for _, frame := range existingFrames {
		if frame == nil || frame.Position <= 0 {
			continue
		}
		occupiedPositions[frame.Position] = true
	}
	position = 1
	for occupiedPositions[position] {
		position++
	}

	if frameModel.IsFrameWithSides(frameType) {
		leftSide := &model.FrameSide{
			Db:     r.Resolver.Db,
			UserID: uid,
		}
		rightSide := &model.FrameSide{
			Db:     r.Resolver.Db,
			UserID: uid,
		}

		leftID, err := leftSide.CreateSide(leftSide)

		if err != nil {
			return nil, err
		}

		rightID, err := rightSide.CreateSide(rightSide)

		if err != nil {
			return nil, err
		}

		frameId, err := frameModel.Create(&boxID, position, frameType, leftID, rightID)

		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return nil, err
		}

		return frameModel.Get(*frameId)

	} else {
		frameId, err := frameModel.Create(&boxID, position, frameType, nil, nil)

		if err != nil {
			logger.ErrorWithContext(ctx, err.Error())
			return nil, err
		}

		return frameModel.Get(*frameId)
	}
}

// UpdateFrames is the resolver for the updateFrames field.
func (r *mutationResolver) UpdateFrames(ctx context.Context, frames []*model.FrameInput) ([]*model.Frame, error) {
	uid := ctx.Value("userID").(string)
	frameModel := &model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	// Initialize an empty results slice
	results := []*model.Frame{}

	for _, frame := range frames {
		if _, err := frameModel.Update(frame.ID, frame.BoxID, frame.Position); err != nil {
			return nil, err
		}
		id, err := strconv.ParseInt(frame.ID, 10, 64)
		if err != nil {
			return nil, err
		}

		updatedFrame, err := frameModel.Get(id)
		if err != nil {
			return nil, err
		}

		results = append(results, updatedFrame)
	}

	return results, nil
}

// DeactivateFrame is the resolver for the deactivateFrame field.
func (r *mutationResolver) DeactivateFrame(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}
