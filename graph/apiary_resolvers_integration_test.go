//go:build integration
// +build integration

package graph

import (
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/require"
)

func TestApiaryFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()

		t.Run("Hives", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.apiary.Hives(fx.ctx, &model.Apiary{ID: fx.apiaryID}, nil, nil)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})
}
