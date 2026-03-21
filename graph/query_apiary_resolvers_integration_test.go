//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryApiaryResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()

		t.Run("Apiaries", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.Apiaries(fx.ctx)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("Apiary", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.Apiary(fx.ctx, strconv.Itoa(fx.apiaryID))

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("ApiaryObstacles", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.ApiaryObstacles(fx.ctx, strconv.Itoa(fx.apiaryID))

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})
}
