package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// Devices is the resolver for the devices field.
func (r *queryResolver) Devices(ctx context.Context) ([]*model.Device, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Device{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).List()
}
