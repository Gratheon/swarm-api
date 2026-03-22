package graph

import (
	"context"
	"strconv"

	"github.com/Gratheon/log-lib-go"
	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/Gratheon/swarm-api/redisPubSub"
)

// AddApiary is the resolver for the addApiary field.
func (r *mutationResolver) AddApiary(ctx context.Context, apiary model.ApiaryInput) (*model.Apiary, error) {
	uid := ctx.Value("userID").(string)
	createdApiary, err := (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(apiary)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	redisPubSub.PublishEvent(uid, "apiary", strconv.Itoa(createdApiary.ID), "created", createdApiary)

	return createdApiary, err
}

// UpdateApiary is the resolver for the updateApiary field.
func (r *mutationResolver) UpdateApiary(ctx context.Context, id string, apiary model.ApiaryInput) (*model.Apiary, error) {
	uid := ctx.Value("userID").(string)
	updatedApiary, err := (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Update(id, apiary)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	redisPubSub.PublishEvent(uid, "apiary", id, "updated", updatedApiary)

	return updatedApiary, err
}

// DeactivateApiary is the resolver for the deactivateApiary field.
func (r *mutationResolver) DeactivateApiary(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	result, err := (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)

	redisPubSub.PublishEvent(uid, "apiary", id, "deleted", "")
	return result, err
}

// UpdateHivePlacement is the resolver for the updateHivePlacement field.
func (r *mutationResolver) UpdateHivePlacement(ctx context.Context, apiaryID string, hiveID string, x float64, y float64, rotation float64) (*model.HivePlacement, error) {
	uid := ctx.Value("userID").(string)
	placement, err := (&model.HivePlacement{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Update(apiaryID, hiveID, x, y, rotation)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return placement, nil
}

// AddApiaryObstacle is the resolver for the addApiaryObstacle field.
func (r *mutationResolver) AddApiaryObstacle(ctx context.Context, apiaryID string, obstacle model.ApiaryObstacleInput) (*model.ApiaryObstacle, error) {
	uid := ctx.Value("userID").(string)
	created, err := (&model.ApiaryObstacle{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(apiaryID, string(obstacle.Type), obstacle.X, obstacle.Y, obstacle.Width, obstacle.Height, obstacle.Radius, obstacle.Rotation, obstacle.Label)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return created, nil
}

// UpdateApiaryObstacle is the resolver for the updateApiaryObstacle field.
func (r *mutationResolver) UpdateApiaryObstacle(ctx context.Context, id string, obstacle model.ApiaryObstacleInput) (*model.ApiaryObstacle, error) {
	uid := ctx.Value("userID").(string)
	updated, err := (&model.ApiaryObstacle{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Update(id, string(obstacle.Type), obstacle.X, obstacle.Y, obstacle.Width, obstacle.Height, obstacle.Radius, obstacle.Rotation, obstacle.Label)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return updated, nil
}

// DeleteApiaryObstacle is the resolver for the deleteApiaryObstacle field.
func (r *mutationResolver) DeleteApiaryObstacle(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	deleted, err := (&model.ApiaryObstacle{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Delete(id)

	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return &deleted, nil
}
