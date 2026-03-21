package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// CreateBoxSystem is the resolver for the createBoxSystem field.
func (r *mutationResolver) CreateBoxSystem(ctx context.Context, name string) (*model.BoxSystem, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystem{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(name)
}

// RenameBoxSystem is the resolver for the renameBoxSystem field.
func (r *mutationResolver) RenameBoxSystem(ctx context.Context, id string, name string) (*model.BoxSystem, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystem{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Rename(id, name)
}

// DeactivateBoxSystem is the resolver for the deactivateBoxSystem field.
func (r *mutationResolver) DeactivateBoxSystem(ctx context.Context, id string, replacementSystemID *string) (bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystem{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id, replacementSystemID)
}

// SetBoxSystemBoxProfileSource is the resolver for the setBoxSystemBoxProfileSource field.
func (r *mutationResolver) SetBoxSystemBoxProfileSource(ctx context.Context, systemID string, boxSourceSystemID *string) (bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystem{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).SetBoxProfileSource(systemID, boxSourceSystemID)
}

// SetBoxSystemFrameSource is the resolver for the setBoxSystemFrameSource field.
func (r *mutationResolver) SetBoxSystemFrameSource(ctx context.Context, systemID string, boxType model.BoxType, frameSourceSystemID string) (bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystemFrameSetting{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).SetFrameSource(systemID, boxType, frameSourceSystemID)
}

// SetBoxSpecDimensions is the resolver for the setBoxSpecDimensions field.
func (r *mutationResolver) SetBoxSpecDimensions(ctx context.Context, systemID string, boxType model.BoxType, internalWidthMm *int, internalLengthMm *int, internalHeightMm *int, externalWidthMm *int, externalLengthMm *int, frameWidthMm *int, frameHeightMm *int) (bool, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSpec{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).SetDimensionsBySystemAndType(
		systemID,
		boxType,
		internalWidthMm,
		internalLengthMm,
		internalHeightMm,
		externalWidthMm,
		externalLengthMm,
		frameWidthMm,
		frameHeightMm,
	)
}
