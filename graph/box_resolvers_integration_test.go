//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/require"
)

func TestBoxFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Box", func(t *testing.T) {
		t.Parallel()

		t.Run("Frames", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.box.Frames(fx.ctx, &model.Box{ID: ptr(strconv.Itoa(fx.boxID))})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})
}
