//go:build integration
// +build integration

package graph

import (
	"strconv"
	"testing"
	"time"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHiveFieldResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Hive", func(t *testing.T) {
		t.Parallel()

		t.Run("Boxes", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.hive.Boxes(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("Family", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.hive.Family(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("Families", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			items, err := fx.hive.Families(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotEmpty(t, items)
		})

		t.Run("BoxCount", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			count, err := fx.hive.BoxCount(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			assert.GreaterOrEqual(t, count, 1)
		})

		t.Run("InspectionCount", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			count, err := fx.hive.InspectionCount(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			assert.GreaterOrEqual(t, count, 1)
		})

		t.Run("LastInspection", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)

			// ACT
			item, err := fx.hive.LastInspection(fx.ctx, &model.Hive{ID: strconv.Itoa(fx.hiveID)})

			// ASSERT
			require.NoError(t, err)
			require.NotNil(t, item)
		})

		t.Run("HiveType", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)

			// ACT
			defaultType, defaultErr := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: ""})
			invalidType, invalidErr := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: "bad"})
			validType, validErr := fx.hive.HiveType(fx.ctx, &model.Hive{HiveType: "HORIZONTAL"})

			// ASSERT
			require.NoError(t, defaultErr)
			require.NoError(t, invalidErr)
			require.NoError(t, validErr)
			assert.Equal(t, model.HiveTypeVertical, defaultType)
			assert.Equal(t, model.HiveTypeVertical, invalidType)
			assert.Equal(t, model.HiveTypeHorizontal, validType)
		})

		t.Run("IsNew", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)
			invalidDate := "not-a-date"
			freshDate := time.Now().Format(time.DateTime)

			// ACT
			isNewWithoutDate, noDateErr := fx.hive.IsNew(fx.ctx, &model.Hive{Added: nil})
			_, invalidErr := fx.hive.IsNew(fx.ctx, &model.Hive{Added: &invalidDate})
			isNewWithFreshDate, freshDateErr := fx.hive.IsNew(fx.ctx, &model.Hive{Added: &freshDate})

			// ASSERT
			require.NoError(t, noDateErr)
			require.Error(t, invalidErr)
			require.NoError(t, freshDateErr)
			assert.False(t, isNewWithoutDate)
			assert.True(t, isNewWithFreshDate)
		})
	})
}
