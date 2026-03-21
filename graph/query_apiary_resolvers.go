package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// Apiary is the resolver for the apiary field.
func (r *queryResolver) Apiary(ctx context.Context, id string) (*model.Apiary, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(id)
}

// Apiaries is the resolver for the apiaries field.
func (r *queryResolver) Apiaries(ctx context.Context) ([]*model.Apiary, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Apiary{
		Db:     r.Db,
		UserID: uid,
	}).List()
}

// HivePlacements is the resolver for the hivePlacements field.
func (r *queryResolver) HivePlacements(ctx context.Context, apiaryID string) ([]*model.HivePlacement, error) {
	uid := ctx.Value("userID").(string)
	return (&model.HivePlacement{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByApiary(apiaryID)
}

// ApiaryObstacles is the resolver for the apiaryObstacles field.
func (r *queryResolver) ApiaryObstacles(ctx context.Context, apiaryID string) ([]*model.ApiaryObstacle, error) {
	uid := ctx.Value("userID").(string)
	return (&model.ApiaryObstacle{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByApiary(apiaryID)
}
