package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// BoxSystems is the resolver for the boxSystems field.
func (r *queryResolver) BoxSystems(ctx context.Context) ([]*model.BoxSystem, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystem{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListVisible()
}

// FrameSpecs is the resolver for the frameSpecs field.
func (r *queryResolver) FrameSpecs(ctx context.Context, systemID *string) ([]*model.FrameSpec, error) {
	uid := ctx.Value("userID").(string)
	return (&model.FrameSpec{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListVisible(systemID)
}

// BoxSpecs is the resolver for the boxSpecs field.
func (r *queryResolver) BoxSpecs(ctx context.Context, systemID string) ([]*model.BoxSpec, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSpec{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListVisible(systemID)
}

// BoxSystemFrameSettings is the resolver for the boxSystemFrameSettings field.
func (r *queryResolver) BoxSystemFrameSettings(ctx context.Context) ([]*model.BoxSystemFrameSetting, error) {
	uid := ctx.Value("userID").(string)
	return (&model.BoxSystemFrameSetting{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).ListAllVisible()
}
