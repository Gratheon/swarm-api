package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// Hive is the resolver for the hive field.
func (r *queryResolver) Hive(ctx context.Context, id string) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(id)
}

// WarehouseQueens is the resolver for the warehouseQueens field.
func (r *queryResolver) WarehouseQueens(ctx context.Context) ([]*model.Family, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListUnassigned()
}
