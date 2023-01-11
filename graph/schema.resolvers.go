package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"gitlab.com/gratheon/swarm-api/graph/generated"
	"gitlab.com/gratheon/swarm-api/graph/model"
	"gitlab.com/gratheon/swarm-api/logger"
)

// Hives is the resolver for the hives field.
func (r *apiaryResolver) Hives(ctx context.Context, obj *model.Apiary) ([]*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByApiary(obj.ID)
}

// Frames is the resolver for the frames field.
func (r *boxResolver) Frames(ctx context.Context, obj *model.Box) ([]*model.Frame, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByBox(obj.ID)
}

// LeftSide is the resolver for the leftSide field.
func (r *frameResolver) LeftSide(ctx context.Context, obj *model.Frame) (*model.FrameSide, error) {
	uid := ctx.Value("userID").(string)
	return (&model.FrameSide{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(obj.LeftID)
}

// RightSide is the resolver for the rightSide field.
func (r *frameResolver) RightSide(ctx context.Context, obj *model.Frame) (*model.FrameSide, error) {
	uid := ctx.Value("userID").(string)
	return (&model.FrameSide{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(obj.RightID)
}

// WorkerCount is the resolver for the workerCount field.
func (r *frameSideResolver) WorkerCount(ctx context.Context, obj *model.FrameSide) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

// DroneCount is the resolver for the droneCount field.
func (r *frameSideResolver) DroneCount(ctx context.Context, obj *model.FrameSide) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

// BoxCount is the resolver for the boxCount field.
func (r *hiveResolver) BoxCount(ctx context.Context, obj *model.Hive) (int, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Count(obj.ID)
}

// Boxes is the resolver for the boxes field.
func (r *hiveResolver) Boxes(ctx context.Context, obj *model.Hive) ([]*model.Box, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(obj.ID)
}

// Family is the resolver for the family field.
func (r *hiveResolver) Family(ctx context.Context, obj *model.Hive) (*model.Family, error) {
	uid := ctx.Value("userID").(string)
	if obj.FamilyID == nil {
		return nil, nil
	}

	return (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).GetById(obj.FamilyID)
}

// Inspections is the resolver for the inspections field.
func (r *hiveResolver) Inspections(ctx context.Context, obj *model.Hive, limit *int) ([]*model.Inspection, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHiveId(obj.ID)
}

// AddApiary is the resolver for the addApiary field.
func (r *mutationResolver) AddApiary(ctx context.Context, apiary model.ApiaryInput) (*model.Apiary, error) {
	uid := ctx.Value("userID").(string)
	createdApiary, err := (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(apiary)

	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	PublishEvent(uid+".apiary", createdApiary)

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
		logger.LogError(err)
		return nil, err
	}

	PublishEvent(uid+".apiary", &updatedApiary)

	return updatedApiary, err
}

// DeactivateApiary is the resolver for the deactivateApiary field.
func (r *mutationResolver) DeactivateApiary(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}

// AddHive is the resolver for the addHive field.
func (r *mutationResolver) AddHive(ctx context.Context, hive model.HiveInput) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	hiveResult, err := (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(hive)

	if err != nil {
		logger.LogError(err)
	}

	err = (&model.Box{
		Db:     r.Db,
		UserID: uid,
	}).CreateByHiveId(hiveResult.ID, hive.BoxCount, hive.Colors)

	if err != nil {
		logger.LogError(err)
	}

	boxes, err := (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(hiveResult.ID)

	if err == nil {
		for _, box := range boxes {
			err = (&model.Frame{
				Db:     r.Db,
				UserID: uid,
			}).CreateFramesForBox(box.ID, hive.FrameCount)
		}
	}

	if err != nil {
		logger.LogError(err)
	}

	return hiveResult, err
}

// UpdateHive is the resolver for the updateHive field.
func (r *mutationResolver) UpdateHive(ctx context.Context, hive model.HiveUpdateInput) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyID := _upsertFamily(r.Resolver.Db, uid, hive)

	err := hiveModel.Update(hive.ID, hive.Name, hive.Notes, familyID)

	if err != nil {
		logger.LogError(err)
	}

	return hiveModel.Get(hive.ID)
}

// DeactivateHive is the resolver for the deactivateHive field.
func (r *mutationResolver) DeactivateHive(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}

// AddBox is the resolver for the addBox field.
func (r *mutationResolver) AddBox(ctx context.Context, hiveID string, position int, color *string, typeArg model.BoxType) (*model.Box, error) {
	uid := ctx.Value("userID").(string)

	boxModel := &model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	boxID, err := boxModel.Create(hiveID, position, color, typeArg)

	if err != nil {
		logger.LogError(err)
	}

	return boxModel.Get(*boxID)
}

// DeactivateBox is the resolver for the deactivateBox field.
func (r *mutationResolver) DeactivateBox(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}

// SwapBoxPositions is the resolver for the swapBoxPositions field.
func (r *mutationResolver) SwapBoxPositions(ctx context.Context, id string, id2 string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).SwapBoxPositions(id, id2)
}

// AddFrame is the resolver for the addFrame field.
func (r *mutationResolver) AddFrame(ctx context.Context, boxID string, typeArg string, position int) (*model.Frame, error) {
	uid := ctx.Value("userID").(string)

	frameModel := &model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	if frameModel.IsFrameWithSides(model.FrameType(typeArg)) {
		leftSide := &model.FrameSide{
			Db:     r.Resolver.Db,
			UserID: uid,
		}
		rightSide := &model.FrameSide{
			Db:     r.Resolver.Db,
			UserID: uid,
		}

		leftID, err := leftSide.CreateSide(leftSide)

		if err != nil {
			return nil, err
		}

		rightID, err := rightSide.CreateSide(rightSide)

		if err != nil {
			return nil, err
		}

		frameId, err := frameModel.Create(&boxID, position, model.FrameType(typeArg), leftID, rightID)

		if err != nil {
			logger.LogError(err)
		}

		return frameModel.Get(*frameId)
	} else {
		frameId, err := frameModel.Create(&boxID, position, model.FrameType(typeArg), nil, nil)

		if err != nil {
			logger.LogError(err)
		}

		return frameModel.Get(*frameId)
	}
}

// DeactivateFrame is the resolver for the deactivateFrame field.
func (r *mutationResolver) DeactivateFrame(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
}

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

// Hive is the resolver for the hive field.
func (r *queryResolver) Hive(ctx context.Context, id string) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(id)
}

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

// Inspection is the resolver for the inspection field.
func (r *queryResolver) Inspection(ctx context.Context, inspectionID string) (*model.Inspection, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get(inspectionID)
}

// Apiary returns generated.ApiaryResolver implementation.
func (r *Resolver) Apiary() generated.ApiaryResolver { return &apiaryResolver{r} }

// Box returns generated.BoxResolver implementation.
func (r *Resolver) Box() generated.BoxResolver { return &boxResolver{r} }

// Frame returns generated.FrameResolver implementation.
func (r *Resolver) Frame() generated.FrameResolver { return &frameResolver{r} }

// FrameSide returns generated.FrameSideResolver implementation.
func (r *Resolver) FrameSide() generated.FrameSideResolver { return &frameSideResolver{r} }

// Hive returns generated.HiveResolver implementation.
func (r *Resolver) Hive() generated.HiveResolver { return &hiveResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type apiaryResolver struct{ *Resolver }
type boxResolver struct{ *Resolver }
type frameResolver struct{ *Resolver }
type frameSideResolver struct{ *Resolver }
type hiveResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *apiaryResolver) Lat(ctx context.Context, obj *model.Apiary) (*string, error) {
	return obj.Lat, nil
}
func (r *apiaryResolver) Lng(ctx context.Context, obj *model.Apiary) (*string, error) {
	return obj.Lng, nil
}
