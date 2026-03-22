package graph

import (
	"context"

	"github.com/Gratheon/log-lib-go"
	"github.com/Gratheon/swarm-api/graph/model"
)

// SetWarehouseModuleCount is the resolver for the setWarehouseModuleCount field.
func (r *mutationResolver) SetWarehouseModuleCount(ctx context.Context, moduleType model.WarehouseModuleType, count int) (*model.WarehouseModule, error) {
	uid := ctx.Value("userID").(string)
	updated, err := (&model.WarehouseModule{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Upsert(moduleType, count)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return updated, nil
}

// SetWarehouseInventoryCount is the resolver for the setWarehouseInventoryCount field.
func (r *mutationResolver) SetWarehouseInventoryCount(ctx context.Context, itemKey string, count int) (*model.WarehouseInventoryItem, error) {
	uid := ctx.Value("userID").(string)
	return (&model.WarehouseInventory{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).UpsertByKey(itemKey, count)
}

// AdjustWarehouseFrameInventory is the resolver for the adjustWarehouseFrameInventory field.
func (r *mutationResolver) AdjustWarehouseFrameInventory(ctx context.Context, boxID string, frameType model.FrameType, delta int) (*model.WarehouseInventoryItem, error) {
	uid := ctx.Value("userID").(string)
	inv := &model.WarehouseInventory{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	specID, err := inv.ResolveFrameSpecIDByBoxAndFrameType(boxID, frameType)
	if err != nil {
		return nil, err
	}
	return inv.UpdateFrameSpecByDelta(specID, delta)
}

// AdjustWarehouseFrameInventoryByFrame is the resolver for the adjustWarehouseFrameInventoryByFrame field.
func (r *mutationResolver) AdjustWarehouseFrameInventoryByFrame(ctx context.Context, frameID string, delta int) (*model.WarehouseInventoryItem, error) {
	uid := ctx.Value("userID").(string)
	inv := &model.WarehouseInventory{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	specID, err := inv.ResolveFrameSpecIDByFrameID(frameID)
	if err != nil {
		return nil, err
	}
	return inv.UpdateFrameSpecByDelta(specID, delta)
}

// SetWarehouseAutoUpdateFromHives is the resolver for the setWarehouseAutoUpdateFromHives field.
func (r *mutationResolver) SetWarehouseAutoUpdateFromHives(ctx context.Context, enabled bool) (*model.WarehouseSettings, error) {
	uid := ctx.Value("userID").(string)
	updated, err := (&model.WarehouseSettings{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).UpsertAutoUpdate(enabled)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return updated, nil
}
