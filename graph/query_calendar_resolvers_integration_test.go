//go:build integration
// +build integration

package graph

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryCalendarResolvers(t *testing.T) {
	t.Parallel()

	t.Run("CalendarAggregate", func(t *testing.T) {
		t.Parallel()

		fx := newSchemaResolverFixture(t, true)
		otherUserID := createTestUserID()
		t.Cleanup(func() {
			cleanupTestData(t, fx.resolver.Db, otherUserID)
		})

		otherApiaryID := createTestApiary(t, fx.resolver.Db, otherUserID)
		otherHiveID := createTestHive(t, fx.resolver.Db, otherUserID, otherApiaryID)
		otherInspectionID, err := (&model.Inspection{Db: fx.resolver.Db, UserID: otherUserID}).Create("{}", otherHiveID)
		require.NoError(t, err)

		oldInspectionDate := time.Now().UTC().AddDate(0, -3, 0).Format("2006-01-02 15:04:05")
		_, err = fx.resolver.Db.Exec(
			"UPDATE inspections SET added=? WHERE id=? AND user_id=?",
			oldInspectionDate,
			fx.inspectionID,
			fx.userID,
		)
		require.NoError(t, err)

		insideDate := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
		inspectionResult := fx.resolver.Db.MustExec(
			"INSERT INTO inspections (user_id, hive_id, data, added) VALUES (?, ?, '{}', ?)",
			fx.userID,
			fx.hiveID,
			insideDate,
		)
		insideInspectionID, err := inspectionResult.LastInsertId()
		require.NoError(t, err)

		hiveLogDetails := "Moved brood frames"
		hiveLog, err := (&model.HiveLog{Db: fx.resolver.Db, UserID: fx.userID}).Create(model.HiveLogInput{
			HiveID:  strconv.Itoa(fx.hiveID),
			Action:  "MANUAL_NOTE",
			Title:   "Calendar hive log",
			Details: &hiveLogDetails,
		})
		require.NoError(t, err)
		_, err = fx.resolver.Db.Exec(
			"UPDATE hive_logs SET created_at=? WHERE id=? AND user_id=?",
			insideDate,
			hiveLog.ID,
			fx.userID,
		)
		require.NoError(t, err)

		from := time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339)
		to := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)

		payload, err := fx.query.Calendar(fx.ctx, model.CalendarInput{
			From: from,
			To:   to,
		})
		require.NoError(t, err)
		require.NotNil(t, payload)

		assert.False(t, payload.Range.Capped)
		assertCalendarItem(t, payload.Items, "inspection:"+strconv.FormatInt(insideInspectionID, 10), model.CalendarItemSourceTypeInspection, strconv.Itoa(fx.hiveID), strconv.Itoa(fx.apiaryID))
		assertCalendarItem(t, payload.Items, "hive-log:"+hiveLog.ID, model.CalendarItemSourceTypeHiveLog, strconv.Itoa(fx.hiveID), strconv.Itoa(fx.apiaryID))
		assertNoCalendarItem(t, payload.Items, "inspection:"+fx.inspectionID)
		assertNoCalendarItem(t, payload.Items, "inspection:"+*otherInspectionID)

		require.NotEmpty(t, payload.InspectionRecency)
		assert.Equal(t, strconv.FormatInt(insideInspectionID, 10), payload.InspectionRecency[0].LatestInspection.ID)
		assert.True(t, payload.InspectionRecency[0].IsInsideSelectedRange)

		filtered, err := fx.query.Calendar(fx.ctx, model.CalendarInput{
			From:        from,
			To:          to,
			SourceTypes: []model.CalendarItemSourceType{model.CalendarItemSourceTypeHiveLog},
		})
		require.NoError(t, err)
		require.NotEmpty(t, filtered.Items)
		for _, item := range filtered.Items {
			assert.Equal(t, model.CalendarItemSourceTypeHiveLog, item.SourceType)
		}
		apiaryFiltered, err := fx.query.Calendar(fx.ctx, model.CalendarInput{
			From:     from,
			To:       to,
			ApiaryID: ptr(strconv.Itoa(fx.apiaryID)),
		})
		require.NoError(t, err)
		require.NotEmpty(t, apiaryFiltered.Items)
		for _, item := range apiaryFiltered.Items {
			assert.Equal(t, strconv.Itoa(fx.apiaryID), strconv.Itoa(item.Apiary.ID))
		}
		wide, err := fx.query.Calendar(fx.ctx, model.CalendarInput{
			From: time.Now().UTC().AddDate(-2, 0, 0).Format(time.RFC3339),
			To:   time.Now().UTC().AddDate(2, 0, 0).Format(time.RFC3339),
		})
		require.NoError(t, err)
		assert.True(t, wide.Range.Capped)
	})

	t.Run("CalendarRejectsOtherUsersHiveFilter", func(t *testing.T) {
		t.Parallel()

		fx := newSchemaResolverFixture(t, true)
		otherCtx := context.WithValue(context.Background(), "userID", createTestUserID())

		payload, err := fx.query.Calendar(otherCtx, model.CalendarInput{
			From:   time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339),
			To:     time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339),
			HiveID: ptr(strconv.Itoa(fx.hiveID)),
		})

		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Empty(t, payload.Items)
		assert.Empty(t, payload.InspectionRecency)
	})
}

func assertCalendarItem(t *testing.T, items []*model.CalendarItem, id string, sourceType model.CalendarItemSourceType, hiveID string, apiaryID string) {
	t.Helper()
	for _, item := range items {
		if item.ID != id {
			continue
		}
		assert.Equal(t, model.CalendarItemKindHistoricalRecord, item.Kind)
		assert.Equal(t, sourceType, item.SourceType)
		require.NotNil(t, item.Source)
		assert.Equal(t, sourceType, item.Source.SourceType)
		require.NotNil(t, item.Source.SourceID)
		assert.Equal(t, idSourceID(id), *item.Source.SourceID)
		require.NotNil(t, item.Source.HiveID)
		assert.Equal(t, hiveID, *item.Source.HiveID)
		require.NotNil(t, item.Source.ApiaryID)
		assert.Equal(t, apiaryID, *item.Source.ApiaryID)
		require.NotNil(t, item.Hive)
		assert.Equal(t, hiveID, item.Hive.ID)
		require.NotNil(t, item.Apiary)
		assert.Equal(t, apiaryID, strconv.Itoa(item.Apiary.ID))
		return
	}
	require.Fail(t, fmt.Sprintf("calendar item %s not found", id))
}

func assertNoCalendarItem(t *testing.T, items []*model.CalendarItem, id string) {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			require.Fail(t, fmt.Sprintf("calendar item %s should not be present", id))
		}
	}
}

func idSourceID(id string) string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == ':' {
			return id[i+1:]
		}
	}
	return id
}
