package graph

import (
	"context"

	"github.com/Gratheon/log-lib-go"
	"github.com/Gratheon/swarm-api/graph/model"
)

// AddDevice is the resolver for the addDevice field.
func (r *mutationResolver) AddDevice(ctx context.Context, device model.DeviceInput) (*model.Device, error) {
	uid := ctx.Value("userID").(string)
	created, err := (&model.Device{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Create(device)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return created, nil
}

// UpdateDevice is the resolver for the updateDevice field.
func (r *mutationResolver) UpdateDevice(ctx context.Context, id string, device model.DeviceUpdateInput) (*model.Device, error) {
	uid := ctx.Value("userID").(string)
	updated, err := (&model.Device{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Update(id, device)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return updated, nil
}

// DeactivateDevice is the resolver for the deactivateDevice field.
func (r *mutationResolver) DeactivateDevice(ctx context.Context, id string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	deactivated, err := (&model.Device{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).Deactivate(id)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return deactivated, nil
}
