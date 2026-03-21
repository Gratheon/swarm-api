//go:build integration
// +build integration

package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolverFactory(t *testing.T) {
	t.Parallel()

	t.Run("ResolverFactory", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		fx := newSchemaResolverFixture(t, false)

		// ACT
		apiaryResolver := fx.resolver.Apiary()
		apiaryObstacleResolver := fx.resolver.ApiaryObstacle()
		boxResolver := fx.resolver.Box()
		familyResolver := fx.resolver.Family()
		frameResolver := fx.resolver.Frame()
		hiveResolver := fx.resolver.Hive()
		mutationResolver := fx.resolver.Mutation()
		queryResolver := fx.resolver.Query()

		// ASSERT
		assert.NotNil(t, apiaryResolver)
		assert.NotNil(t, apiaryObstacleResolver)
		assert.NotNil(t, boxResolver)
		assert.NotNil(t, familyResolver)
		assert.NotNil(t, frameResolver)
		assert.NotNil(t, hiveResolver)
		assert.NotNil(t, mutationResolver)
		assert.NotNil(t, queryResolver)
	})
}
