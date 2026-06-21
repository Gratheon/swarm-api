package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// Calendar is the resolver for the calendar field.
func (r *queryResolver) Calendar(ctx context.Context, input model.CalendarInput) (*model.CalendarPayload, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Calendar{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Payload(input)
}
