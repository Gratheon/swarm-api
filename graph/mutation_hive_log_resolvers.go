package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// AddHiveLog is the resolver for the addHiveLog field.
func (r *mutationResolver) AddHiveLog(ctx context.Context, log model.HiveLogInput) (*model.HiveLog, error) {
	uid := ctx.Value("userID").(string)
	return (&model.HiveLog{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(log)
}

// UpdateHiveLog is the resolver for the updateHiveLog field.
func (r *mutationResolver) UpdateHiveLog(ctx context.Context, id string, log model.HiveLogUpdateInput) (*model.HiveLog, error) {
	uid := ctx.Value("userID").(string)
	return (&model.HiveLog{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Update(id, log)
}

// DeleteHiveLog is the resolver for the deleteHiveLog field.
func (r *mutationResolver) DeleteHiveLog(ctx context.Context, id string) (bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.HiveLog{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Delete(id)
}
