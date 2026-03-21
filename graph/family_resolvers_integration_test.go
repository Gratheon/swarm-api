//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/require"
)

func TestFamilyFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Family", func(t *testing.T) {
		t.Parallel()

		t.Run("LastTreatment", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.family.LastTreatment(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("Treatments", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.family.Treatments(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("LastHive", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.family.LastHive(fx.ctx, &model.Family{ID: strconv.Itoa(fx.familyID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})
	})
}
