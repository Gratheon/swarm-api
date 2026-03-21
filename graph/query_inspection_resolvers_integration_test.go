//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryInspectionResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Inspection", func(t *testing.T) {
		t.Parallel()

		t.Run("Inspection", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.query.Inspection(fx.ctx, fx.inspectionID)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("Inspections", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.query.Inspections(fx.ctx, strconv.Itoa(fx.hiveID), nil)

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})
	})
}
