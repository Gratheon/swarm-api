//go:build integration
// +build integration

package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryNamingResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Naming", func(t *testing.T) {
		t.Parallel()

		t.Run("RandomHiveName", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			name, err := fx.query.RandomHiveName(fx.ctx, nil)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, name)
		})

		t.Run("RandomHiveNameFallbackToBee", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			resolver := &queryResolver{Resolver: &Resolver{femaleNamesMap: map[string][]string{}}}

			// ACT
			name, err := resolver.RandomHiveName(context.Background(), nil)

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, name)
			assert.Equal(t, "Bee", *name)
		})
	})
}
