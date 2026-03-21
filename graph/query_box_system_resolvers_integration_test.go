//go:build integration
// +build integration

package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBoxSystemResolvers(t *testing.T) {
	t.Parallel()

	t.Run("BoxSystem", func(t *testing.T) {
		t.Parallel()

		t.Run("BoxSystems", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.BoxSystems(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("FrameSpecs", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.FrameSpecs(fx.ctx, nil)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("BoxSpecs", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			systems, err := fx.query.BoxSystems(fx.ctx)
			require.NoError(t, err)
			require.NotEmpty(t, systems)

			// ACT
			items, err := fx.query.BoxSpecs(fx.ctx, systems[0].ID)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})

		t.Run("BoxSystemFrameSettings", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.BoxSystemFrameSettings(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})
	})
}
