package graph

import (
	"context"
	"strconv"

	"github.com/Gratheon/swarm-api/graph/model"
)

// HiveFrame is the resolver for the hiveFrame field.
func (r *queryResolver) HiveFrame(ctx context.Context, id string) (*model.Frame, error) {
	uid := ctx.Value("userID").(string)
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}
	return (&model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(idNum)
}

// HiveFrameSide is the resolver for the hiveFrameSide field.
func (r *queryResolver) HiveFrameSide(ctx context.Context, id string) (*model.FrameSide, error) {
	uid := ctx.Value("userID").(string)
	idNum, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return nil, err
	}
	idNum2 := int(idNum)

	return (&model.FrameSide{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(&idNum2)
}
