//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryFrameResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Frame", func(t *testing.T) {
		t.Parallel()

		t.Run("HiveFrame", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			_, invalidErr := fx.query.HiveFrame(fx.ctx, "not-a-number")
			item, err := fx.query.HiveFrame(fx.ctx, strconv.Itoa(fx.frameID))

			// ASSERT
			require.Error(t, invalidErr)
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("HiveFrameSide", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			_, invalidErr := fx.query.HiveFrameSide(fx.ctx, "invalid")
			item, err := fx.query.HiveFrameSide(fx.ctx, strconv.Itoa(fx.leftSideID))

			// ASSERT
			require.Error(t, invalidErr)
			require.NoError(t, err)
			require.NotNil(t, item)
		})
	})
}
