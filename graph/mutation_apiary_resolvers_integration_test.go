//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutationApiaryResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Apiary", func(t *testing.T) {
		t.Parallel()

		t.Run("Add", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)

			// ACT
			created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "Resolver Apiary"})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, created)
		})

		t.Run("Update", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "Before"})
			require.NoError(t, err)

			// ACT
			updated, err := fx.mutation.UpdateApiary(fx.ctx, strconv.Itoa(created.ID), model.ApiaryInput{Name: "After"})

			// ASSERT
			require.NoError(t, err)
			assert.Equal(t, "After", *updated.Name)
		})

		t.Run("Deactivate", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			created, err := fx.mutation.AddApiary(fx.ctx, model.ApiaryInput{Name: "ToDeactivate"})
			require.NoError(t, err)

			// ACT
			ok, err := fx.mutation.DeactivateApiary(fx.ctx, strconv.Itoa(created.ID))

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, ok)
			assert.True(t, *ok)
		})
	})
}
