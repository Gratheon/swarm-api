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

func TestMutationDeviceResolvers(t *testing.T) {
	t.Parallel()

	t.Run("Device", func(t *testing.T) {
		t.Parallel()

		t.Run("AddUpdateDeactivate", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			hiveID := strconv.Itoa(fx.hiveID)
			boxID := strconv.Itoa(fx.boxID)
			token := " token-1 "

			// ACT
			created, createErr := fx.mutation.AddDevice(fx.ctx, model.DeviceInput{
				Name:     " Hive sensor ",
				Type:     model.DeviceTypeIotSensor,
				APIToken: &token,
				HiveID:   &hiveID,
				BoxID:    &boxID,
			})
			newName := "Updated sensor"
			newType := model.DeviceTypeVideoCamera
			empty := ""
			updated, updateErr := fx.mutation.UpdateDevice(fx.ctx, created.ID, model.DeviceUpdateInput{
				Name:     &newName,
				Type:     &newType,
				APIToken: &empty,
				HiveID:   &empty,
			})
			deactivated, deactivateErr := fx.mutation.DeactivateDevice(fx.ctx, created.ID)
			items, listErr := fx.query.Devices(fx.ctx)

			// ASSERT
			require.NoError(t, createErr)
			require.NotNil(t, created)
			require.NotNil(t, created.BoxID)
			require.NotNil(t, created.HiveID)
			assert.Equal(t, fx.boxID, *created.BoxID)
			assert.Equal(t, fx.hiveID, *created.HiveID)
			require.NoError(t, updateErr)
			require.NotNil(t, updated)
			assert.Equal(t, "Updated sensor", updated.Name)
			assert.Equal(t, model.DeviceTypeVideoCamera, updated.Type)
			assert.Nil(t, updated.APIToken)
			assert.Nil(t, updated.HiveID)
			assert.Nil(t, updated.BoxID)
			require.NoError(t, deactivateErr)
			require.NotNil(t, deactivated)
			assert.True(t, *deactivated)
			require.NoError(t, listErr)
			assert.False(t, hasDeviceID(items, created.ID))
		})

		t.Run("AddFailsWhenBoxHiveMismatch", func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			fx := newSchemaResolverFixture(t, true)
			otherHiveID := createTestHive(t, fx.resolver.Db, fx.userID, fx.apiaryID)
			otherHiveIDString := strconv.Itoa(otherHiveID)
			boxID := strconv.Itoa(fx.boxID)

			// ACT
			_, err := fx.mutation.AddDevice(fx.ctx, model.DeviceInput{
				Name:   "Mismatch device",
				Type:   model.DeviceTypeIotSensor,
				HiveID: &otherHiveIDString,
				BoxID:  &boxID,
			})

			// ASSERT
			require.Error(t, err)
			assert.ErrorContains(t, err, "selected box does not belong to selected hive")
		})
	})
}
