package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// AddInspection is the resolver for the addInspection field.
func (r *mutationResolver) AddInspection(ctx context.Context, inspection model.InspectionInput) (*model.Inspection, error) {
	uid := ctx.Value("userID").(string)
	inspectionModel := &model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	id, err := inspectionModel.Create(inspection.Data, inspection.HiveID)
	if err != nil {
		return nil, err
	}

	return inspectionModel.Get(*id)
}
