package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"
	"time"

	"github.com/Gratheon/swarm-api/graph/generated"
	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/Gratheon/swarm-api/logger"
	"github.com/Gratheon/swarm-api/redisPubSub"
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

// LastTreatment is the resolver for the lastTreatment field.
func (r *familyResolver) LastTreatment(ctx context.Context, obj *model.Family) (*string, error) {
	uid := ctx.Value("userID").(string)
	treatmentModel := &model.Treatment{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	treatment, err := treatmentModel.GetLastFamilyTreatment(obj.ID)
	if err != nil {
		return nil, err
	}

	if treatment == nil {
		return nil, nil
	}

	return &treatment.Added, nil
}

// Treatments is the resolver for the treatments field.
func (r *familyResolver) Treatments(ctx context.Context, obj *model.Family) ([]*model.Treatment, error) {
	uid := ctx.Value("userID").(string)
	treatmentModel := &model.Treatment{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	return treatmentModel.ListFamilyTreatments(obj.ID)
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

// BoxCount is the resolver for the boxCount field.
func (r *hiveResolver) BoxCount(ctx context.Context, obj *model.Hive) (int, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Count(obj.ID)
}

// InspectionCount is the resolver for the inspectionCount field.
func (r *hiveResolver) InspectionCount(ctx context.Context, obj *model.Hive) (int, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).CountByHiveId(obj.ID)
}

// IsNew is the resolver for the isNew field.
func (r *hiveResolver) IsNew(ctx context.Context, obj *model.Hive) (bool, error) {
	if obj.Added == nil {
		return false, nil
	}
	// return true if obj.Added is less than 1 day old
	now := time.Now()
	added, err := time.Parse(time.DateTime, *obj.Added)

	if err != nil {
		return false, err
	}

	return now.Sub(added).Hours() < 24, nil
}

// LastInspection is the resolver for the lastInspection field.
func (r *hiveResolver) LastInspection(ctx context.Context, obj *model.Hive) (*string, error) {
	// TODO use dataloader to avoid querying for each hive

	uid := ctx.Value("userID").(string)
	inspectionModel := &model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	inspection, err := inspectionModel.GetLatestByHiveId(obj.ID)
	if err != nil {
		return nil, err
	}

	if inspection == nil {
		return nil, nil
	}

	return &inspection.Added, nil
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

	redisPubSub.PublishEvent(uid, "apiary", strconv.Itoa(createdApiary.ID), "created", createdApiary)

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

	redisPubSub.PublishEvent(uid, "apiary", id, "updated", updatedApiary)

	return updatedApiary, err
}

// DeactivateApiary is the resolver for the deactivateApiary field.
func (r *mutationResolver) DeactivateApiary(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	result, err := (&model.Apiary{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)

	redisPubSub.PublishEvent(uid, "apiary", id, "deleted", "")
	return result, err
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

	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyID, err := familyModel.Upsert(uid, hive)
	if err != nil {
		logger.LogError(err)
	}

	err = hiveModel.Update(hive.ID, hive.Name, hive.Notes, familyID)

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

// UpdateBoxColor is the resolver for the updateBoxColor field.
func (r *mutationResolver) UpdateBoxColor(ctx context.Context, id string, color *string) (bool, error) {
	uid := ctx.Value("userID").(string)
	boxModel := &model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	box, err := boxModel.Get(id)

	if err != nil {
		logger.LogError(err)
	}

	box.Color = color

	return boxModel.Update(box.ID, *box.Position, box.Color)
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

// UpdateFrames is the resolver for the updateFrames field.
func (r *mutationResolver) UpdateFrames(ctx context.Context, frames []*model.FrameInput) ([]*model.Frame, error) {
	uid := ctx.Value("userID").(string)
	frameModel := &model.Frame{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	// Initialize an empty results slice
	results := []*model.Frame{}

	for _, frame := range frames {
		frameModel.Update(frame.ID, frame.BoxID, frame.Position)
		id, err := strconv.ParseInt(frame.ID, 10, 64)
		if err != nil {
			return nil, err
		}

		updatedFrame, err := frameModel.Get(id)
		if err != nil {
			return nil, err
		}

		results = append(results, updatedFrame)
	}

	return results, nil
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

// TreatHive is the resolver for the treatHive field.
func (r *mutationResolver) TreatHive(ctx context.Context, treatment model.TreatmentOfHiveInput) (*bool, error) {
	uid := ctx.Value("userID").(string)
	treatmentModel := &model.Treatment{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	hive, err := hiveModel.Get(treatment.HiveID)
	if err != nil {
		logger.LogError(err)
		ok := err == nil
		return &ok, err
	}

	_, err2 := treatmentModel.TreatHive(treatment, hive.FamilyID)
	ok := err2 == nil

	if err2 != nil {
		logger.LogError(err2)
	}

	return &ok, err
}

// TreatBox is the resolver for the treatBox field.
func (r *mutationResolver) TreatBox(ctx context.Context, treatment model.TreatmentOfBoxInput) (*bool, error) {
	uid := ctx.Value("userID").(string)
	treatmentModel := &model.Treatment{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	hiveModel := &model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	hive, err := hiveModel.Get(treatment.HiveID)
	if err != nil {
		logger.LogError(err)
		ok := err == nil
		return &ok, err
	}

	_, err2 := treatmentModel.TreatHiveBox(treatment, hive.FamilyID)
	ok := err2 == nil
	if err2 != nil {
		logger.LogError(err2)
	}
	return &ok, err
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

// Inspections is the resolver for the inspections field.
func (r *queryResolver) Inspections(ctx context.Context, hiveID string, limit *int) ([]*model.Inspection, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Inspection{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHiveId(hiveID)
}

// Apiary returns generated.ApiaryResolver implementation.
func (r *Resolver) Apiary() generated.ApiaryResolver { return &apiaryResolver{r} }

// Box returns generated.BoxResolver implementation.
func (r *Resolver) Box() generated.BoxResolver { return &boxResolver{r} }

// Family returns generated.FamilyResolver implementation.
func (r *Resolver) Family() generated.FamilyResolver { return &familyResolver{r} }

// Frame returns generated.FrameResolver implementation.
func (r *Resolver) Frame() generated.FrameResolver { return &frameResolver{r} }

// Hive returns generated.HiveResolver implementation.
func (r *Resolver) Hive() generated.HiveResolver { return &hiveResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type apiaryResolver struct{ *Resolver }
type boxResolver struct{ *Resolver }
type familyResolver struct{ *Resolver }
type frameResolver struct{ *Resolver }
type hiveResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
