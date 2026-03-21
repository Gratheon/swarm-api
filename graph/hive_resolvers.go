package graph

import (
	"context"
	"strings"
	"time"

	"github.com/Gratheon/swarm-api/graph/model"
)

// HiveType is the resolver for the hiveType field.
func (r *hiveResolver) HiveType(ctx context.Context, obj *model.Hive) (model.HiveType, error) {
	value := strings.ToUpper(strings.TrimSpace(obj.HiveType))
	if value == "" {
		return model.HiveTypeVertical, nil
	}
	hiveType := model.HiveType(value)
	if !hiveType.IsValid() {
		return model.HiveTypeVertical, nil
	}
	return hiveType, nil
}

// Boxes is the resolver for the boxes field.
func (r *hiveResolver) Boxes(ctx context.Context, obj *model.Hive) ([]*model.Box, error) {
	uid := ctx.Value("userID").(string)
	loaders := GetLoaders(ctx)
	if loaders != nil && loaders.BoxesByHiveLoader != nil {
		return loaders.BoxesByHiveLoader.Load(ctx, obj.ID, uid)
	}
	return (&model.Box{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(obj.ID)
}

// Family is the resolver for the family field.
func (r *hiveResolver) Family(ctx context.Context, obj *model.Hive) (*model.Family, error) {
	uid := ctx.Value("userID").(string)

	loaders := GetLoaders(ctx)
	if loaders != nil && loaders.FamilyByHiveLoader != nil {
		return loaders.FamilyByHiveLoader.Load(ctx, obj.ID, uid)
	}

	families, err := (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(obj.ID)

	if err != nil || len(families) == 0 {
		return nil, err
	}

	return families[0], nil
}

// Families is the resolver for the families field.
func (r *hiveResolver) Families(ctx context.Context, obj *model.Hive) ([]*model.Family, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListByHive(obj.ID)
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

// ParentHive is the resolver for the parentHive field.
func (r *hiveResolver) ParentHive(ctx context.Context, obj *model.Hive) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).GetParentHive(obj.ParentHiveID)
}

// ChildHives is the resolver for the childHives field.
func (r *hiveResolver) ChildHives(ctx context.Context, obj *model.Hive) ([]*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).GetChildHives(obj.ID)
}

// MergedIntoHive is the resolver for the mergedIntoHive field.
func (r *hiveResolver) MergedIntoHive(ctx context.Context, obj *model.Hive) (*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).GetMergedIntoHive(obj.MergedIntoHiveID)
}

// MergedFromHives is the resolver for the mergedFromHives field.
func (r *hiveResolver) MergedFromHives(ctx context.Context, obj *model.Hive) ([]*model.Hive, error) {
	uid := ctx.Value("userID").(string)
	return (&model.Hive{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).GetMergedFromHives(obj.ID)
}
