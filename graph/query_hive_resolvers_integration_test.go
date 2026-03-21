//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryHiveResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Hive", func(t *testing.T) {
		t.Parallel()

		t.Run("Hive", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.Hive(fx.ctx, strconv.Itoa(fx.hiveID))

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("HivePlacements", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.HivePlacements(fx.ctx, strconv.Itoa(fx.apiaryID))

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("HiveLogs", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.HiveLogs(fx.ctx, strconv.Itoa(fx.hiveID), nil)

			// ASSERT
			require.NoError(t, err)
			assert.NotNil(t, items)
		})
	})
}
