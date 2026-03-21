//go:build integration
// +build integration

package graph

import (
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryWarehouseResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Warehouse", func(t *testing.T) {
		t.Parallel()

		t.Run("Devices", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.Devices(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("WarehouseModules", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.WarehouseModules(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("WarehouseInventory", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.WarehouseInventory(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("WarehouseSettings", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.WarehouseSettings(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("WarehouseModuleStats", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.WarehouseModuleStats(fx.ctx, model.WarehouseModuleTypeRoof)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("WarehouseInventoryStats", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.WarehouseInventoryStats(fx.ctx, "BOX:ROOF")

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("WarehouseQueens", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.WarehouseQueens(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})
	})
}
