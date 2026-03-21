package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// HiveLogs is the resolver for the hiveLogs field.
func (r *queryResolver) HiveLogs(ctx context.Context, hiveID string, limit *int) ([]*model.HiveLog, error) {
	uid := ctx.Value("userID").(string)
	return (&model.HiveLog{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(hiveID, limit)
}
