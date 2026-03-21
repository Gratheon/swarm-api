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

func TestMutationHiveLogResolvers(t *testing.T) {
	t.Parallel()

	t.Run("HiveLog", func(t *testing.T) {
		t.Parallel()

		t.Run("AddUpdateDelete", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			details := "initial details"
			source := "integration-test"
			dedupeKey := "schema-resolver-hive-log-1"
			relatedHive := &model.HiveLogRelatedHiveInput{ID: strconv.Itoa(fx.hiveID)}

			// ACT
			created, createErr := fx.mutation.AddHiveLog(fx.ctx, model.HiveLogInput{
				HiveID:       strconv.Itoa(fx.hiveID),
				Action:       "INSPECTION",
				Title:        "Initial title",
				Details:      &details,
				Source:       &source,
				DedupeKey:    &dedupeKey,
				RelatedHives: []*model.HiveLogRelatedHiveInput{relatedHive},
			})
			updatedTitle := "Updated title"
			updatedDetails := "updated details"
			updated, updateErr := fx.mutation.UpdateHiveLog(fx.ctx, created.ID, model.HiveLogUpdateInput{
				Title:   &updatedTitle,
				Details: &updatedDetails,
			})
			deleted, deleteErr := fx.mutation.DeleteHiveLog(fx.ctx, created.ID)
			items, listErr := fx.query.HiveLogs(fx.ctx, strconv.Itoa(fx.hiveID), nil)

			// ASSERT
			require.NoError(t, createErr)
			require.NotNil(t, created)
			assert.Equal(t, "Initial title", created.Title)
			assert.Equal(t, "INSPECTION", created.Action)
			require.NoError(t, updateErr)
			require.NotNil(t, updated)
			assert.Equal(t, updatedTitle, updated.Title)
			assert.Equal(t, updatedDetails, *updated.Details)
			require.NoError(t, deleteErr)
			assert.True(t, deleted)
			require.NoError(t, listErr)
			assert.False(t, hasHiveLogID(items, created.ID))
		})

		t.Run("AddFailsForUnknownHive", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, false)

			// ACT
			_, err := fx.mutation.AddHiveLog(fx.ctx, model.HiveLogInput{
				HiveID: "99999999",
				Action: "INSPECTION",
				Title:  "Should fail",
			})

			// ASSERT
			require.Error(t, err)
			assert.ErrorContains(t, err, "hive not found")
		})
	})
}
