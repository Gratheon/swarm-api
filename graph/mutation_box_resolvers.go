package graph

import (
	"context"
	"errors"
	"strings"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/Gratheon/swarm-api/logger"
)

// AddBox is the resolver for the addBox field.
func (r *mutationResolver) AddBox(ctx context.Context, hiveID string, position int, color *string, typeArg model.BoxType, holeCount *int) (*model.Box, error) {
	uid := ctx.Value("userID").(string)

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	hive, err := hiveModel.Get(hiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}
	if hive != nil && strings.EqualFold(hive.HiveType, model.HiveTypeNucleus.String()) {
		return nil, errors.New("nucleus hives are monolithic and do not support adding extra sections")
	}

	boxModel := &model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	boxID, err := boxModel.Create(hiveID, position, color, typeArg, holeCount)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
	}

	return boxModel.Get(*boxID)
}

// UpdateBoxColor is the resolver for the updateBoxColor field.
func (r *mutationResolver) UpdateBoxColor(ctx context.Context, id string, color *string) (bool, error) {
	uid := ctx.Value("userID").(string)
	boxModel := &model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	box, err := boxModel.Get(id)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
	}

	box.Color = color

	return boxModel.Update(box.ID, *box.Position, box.Color)
}

// UpdateBoxHoleCount is the resolver for the updateBoxHoleCount field.
func (r *mutationResolver) UpdateBoxHoleCount(ctx context.Context, id string, holeCount int) (bool, error) {
	uid := ctx.Value("userID").(string)
	boxModel := &model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	box, err := boxModel.Get(id)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return false, err
	}
	if box == nil {
		return false, errors.New("box not found")
	}
	if box.Type != model.BoxTypeGate {
		return false, errors.New("hole count can only be updated for gate boxes")
	}

	return boxModel.UpdateHoleCount(id, holeCount)
}

// UpdateBoxRoofStyle is the resolver for the updateBoxRoofStyle field.
func (r *mutationResolver) UpdateBoxRoofStyle(ctx context.Context, id string, roofStyle model.RoofStyle) (bool, error) {
	uid := ctx.Value("userID").(string)
	boxModel := &model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	box, err := boxModel.Get(id)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return false, err
	}
	if box == nil {
		return false, errors.New("box not found")
	}
	if box.Type != model.BoxTypeRoof {
		return false, errors.New("roof style can only be updated for roof boxes")
	}

	return boxModel.UpdateRoofStyle(id, roofStyle)
}

// DeactivateBox is the resolver for the deactivateBox field.
func (r *mutationResolver) DeactivateBox(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}

// SwapBoxPositions is the resolver for the swapBoxPositions field.
func (r *mutationResolver) SwapBoxPositions(ctx context.Context, id string, id2 string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).SwapBoxPositions(id, id2)
}
