//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutationFamilyResolvers(t *testing.T) {
	t.Parallel()

	t.Run("QueenWarehouse", func(t *testing.T) {
		t.Parallel()

		t.Run("MoveToWarehouseAndAssignBack", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			hiveID := strconv.Itoa(fx.hiveID)
			familyID := strconv.Itoa(fx.familyID)

			// ACT
			moved, moveErr := fx.mutation.MoveQueenToWarehouse(fx.ctx, hiveID, familyID)
			warehouseQueens, warehouseErr := fx.query.WarehouseQueens(fx.ctx)
			assigned, assignErr := fx.mutation.AssignQueenFromWarehouse(fx.ctx, hiveID, familyID)

			// ASSERT
			require.NoError(t, moveErr)
			require.NotNil(t, moved)
			assert.Nil(t, moved.HiveID)
			require.NoError(t, warehouseErr)
			assert.True(t, hasFamilyID(warehouseQueens, familyID))
			require.NoError(t, assignErr)
			require.NotNil(t, assigned)
			require.NotNil(t, assigned.HiveID)
			assert.Equal(t, fx.hiveID, *assigned.HiveID)
		})
	})
}
