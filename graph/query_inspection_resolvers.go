package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// Inspection is the resolver for the inspection field.
func (r *queryResolver) Inspection(ctx context.Context, inspectionID string) (*model.Inspection, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(inspectionID)
}

// Inspections is the resolver for the inspections field.
func (r *queryResolver) Inspections(ctx context.Context, hiveID string, limit *int) ([]*model.Inspection, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHiveId(hiveID)
}
