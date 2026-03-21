package graph

import (
	"context"

	"github.com/Gratheon/swarm-api/graph/model"
)

// WarehouseModules is the resolver for the warehouseModules field.
func (r *queryResolver) WarehouseModules(ctx context.Context) ([]*model.WarehouseModule, error) {
	uid := ctx.Value("userID").(string)
	return (&model.WarehouseModule{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).List()
}

// WarehouseInventory is the resolver for the warehouseInventory field.
func (r *queryResolver) WarehouseInventory(ctx context.Context) ([]*model.WarehouseInventoryItem, error) {
	uid := ctx.Value("userID").(string)
	return (&model.WarehouseInventory{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).List()
}

// WarehouseSettings is the resolver for the warehouseSettings field.
func (r *queryResolver) WarehouseSettings(ctx context.Context) (*model.WarehouseSettings, error) {
	uid := ctx.Value("userID").(string)
	return (&model.WarehouseSettings{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Get()
}

// WarehouseModuleStats is the resolver for the warehouseModuleStats field.
func (r *queryResolver) WarehouseModuleStats(ctx context.Context, moduleType model.WarehouseModuleType) (*model.WarehouseModuleStats, error) {
	uid := ctx.Value("userID").(string)
	return (&model.WarehouseModule{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).UsageStats(moduleType)
}

// WarehouseInventoryStats is the resolver for the warehouseInventoryStats field.
func (r *queryResolver) WarehouseInventoryStats(ctx context.Context, itemKey string) (*model.WarehouseInventoryStats, error) {
	uid := ctx.Value("userID").(string)
	return (&model.WarehouseInventory{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).StatsByKey(itemKey)
}
