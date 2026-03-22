package graph

import (
	"context"
	"strconv"

	"github.com/Gratheon/log-lib-go"
	"github.com/Gratheon/swarm-api/graph/model"
)

// AddQueenToHive is the resolver for the addQueenToHive field.
func (r *mutationResolver) AddQueenToHive(ctx context.Context, hiveID string, queen model.FamilyInput) (*model.Family, error) {
	uid := ctx.Value("userID").(string)
	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyID, err := familyModel.CreateForHive(hiveID, queen.Name, queen.Race, queen.Added, queen.Color)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return familyModel.GetById(familyID)
}

// AddWarehouseQueen is the resolver for the addWarehouseQueen field.
func (r *mutationResolver) AddWarehouseQueen(ctx context.Context, queen model.FamilyInput) (*model.Family, error) {
	uid := ctx.Value("userID").(string)
	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyID, err := familyModel.Create(queen.Name, queen.Race, queen.Added, queen.Color)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return familyModel.GetById(familyID)
}

// RemoveQueenFromHive is the resolver for the removeQueenFromHive field.
func (r *mutationResolver) RemoveQueenFromHive(ctx context.Context, hiveID string, familyID string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	success, err := (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).DeleteFromHive(hiveID, familyID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}
	return &success, nil
}

// TreatHive is the resolver for the treatHive field.
func (r *mutationResolver) TreatHive(ctx context.Context, treatment model.TreatmentOfHiveInput) (*bool, error) {
	uid := ctx.Value("userID").(string)
	treatmentModel := &model.Treatment{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	families, err := familyModel.ListByHive(treatment.HiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		ok := false
		return &ok, err
	}
	var familyID *int
	if err == nil && len(families) > 0 {
		familyIDInt, _ := strconv.Atoi(families[0].ID)
		familyID = &familyIDInt
	}

	_, err2 := treatmentModel.TreatHive(treatment, familyID)
	ok := err2 == nil

	if err2 != nil {
		logger.ErrorWithContext(ctx, err2.Error())
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

	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}
	families, err := familyModel.ListByHive(treatment.HiveID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		ok := false
		return &ok, err
	}
	var familyID *int
	if err == nil && len(families) > 0 {
		familyIDInt, _ := strconv.Atoi(families[0].ID)
		familyID = &familyIDInt
	}

	_, err2 := treatmentModel.TreatHiveBox(treatment, familyID)
	ok := err2 == nil
	if err2 != nil {
		logger.ErrorWithContext(ctx, err2.Error())
	}
	return &ok, err
}

// MoveQueenToWarehouse is the resolver for the moveQueenToWarehouse field.
func (r *mutationResolver) MoveQueenToWarehouse(ctx context.Context, hiveID string, familyID string) (*model.Family, error) {
	uid := ctx.Value("userID").(string)
	moved, err := (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).MoveToWarehouse(hiveID, familyID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return moved, nil
}

// AssignQueenFromWarehouse is the resolver for the assignQueenFromWarehouse field.
func (r *mutationResolver) AssignQueenFromWarehouse(ctx context.Context, hiveID string, familyID string) (*model.Family, error) {
	uid := ctx.Value("userID").(string)
	assigned, err := (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).AssignFromWarehouse(hiveID, familyID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return assigned, nil
}

// DeleteWarehouseQueen is the resolver for the deleteWarehouseQueen field.
func (r *mutationResolver) DeleteWarehouseQueen(ctx context.Context, familyID string) (*bool, error) {
	uid := ctx.Value("userID").(string)
	success, err := (&model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}).DeleteFromWarehouse(familyID)
	if err != nil {
		logger.ErrorWithContext(ctx, err.Error())
		return nil, err
	}

	return &success, nil
}
