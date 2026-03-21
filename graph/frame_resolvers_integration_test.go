//go:build integration
// +build integration

package graph

import (
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/require"
)

func TestFrameFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Frame", func(t *testing.T) {
		t.Parallel()

		t.Run("LeftSide", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.frame.LeftSide(fx.ctx, &model.Frame{LeftID: &fx.leftSideID})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("RightSide", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.frame.RightSide(fx.ctx, &model.Frame{RightID: &fx.rightSideID})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})
	})
}
