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

func requireLangstrothSeeded(t *testing.T, dbModel *model.BoxSystem) {
	t.Helper()

	var seededCount int
	err := dbModel.Db.Get(&seededCount, `
		SELECT COUNT(*)
		FROM box_systems
		WHERE user_id IS NULL
		  AND name = 'Langstroth'
		  AND active = 1
	`)
	require.NoError(t, err)
	if seededCount == 0 {
		t.Skip("Skipping test - Langstroth seed box system is required")
	}
}

func cleanupBoxSystemsByUser(t *testing.T, dbModel *model.BoxSystem) {
	t.Helper()
	cleanupTestData(t, dbModel.Db, dbModel.UserID)
	_, _ = dbModel.Db.Exec("DELETE FROM box_systems WHERE user_id=?", dbModel.UserID)
}

func insertOwnedBoxSystem(t *testing.T, dbModel *model.BoxSystem, name string) string {
	t.Helper()

	result := dbModel.Db.MustExec(
		"INSERT INTO box_systems (user_id, name, is_default, active) VALUES (?, ?, 0, 1)",
		dbModel.UserID,
		name,
	)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return strconv.FormatInt(id, 10)
}

func TestBoxSystemCreateLimit(t *testing.T) {
	t.Parallel()

	dbConn := setupTestDB(t)
	if dbConn == nil {
		return
	}
	defer dbConn.Close()

	userID := createTestUserID()

	systemModel := &model.BoxSystem{
		Db:     dbConn,
		UserID: userID,
	}
	defer cleanupBoxSystemsByUser(t, systemModel)
	requireLangstrothSeeded(t, systemModel)
	for i := 1; i <= 4; i++ {
		insertOwnedBoxSystem(t, systemModel, "Custom system "+strconv.Itoa(i))
	}

	created, err := systemModel.Create(" Fifth system ")
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "Fifth system", created.Name)

	sixth, err := systemModel.Create("Sixth system")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "maximum 5 box systems allowed")
	assert.Nil(t, sixth)
}

func TestSetBoxProfileSourceDetectsCycle(t *testing.T) {
	t.Parallel()

	dbConn := setupTestDB(t)
	if dbConn == nil {
		return
	}
	defer dbConn.Close()

	userID := createTestUserID()

	systemModel := &model.BoxSystem{
		Db:     dbConn,
		UserID: userID,
	}
	defer cleanupBoxSystemsByUser(t, systemModel)
	systemA := insertOwnedBoxSystem(t, systemModel, "System A")
	systemB := insertOwnedBoxSystem(t, systemModel, "System B")
	systemC := insertOwnedBoxSystem(t, systemModel, "System C")

	ok, err := systemModel.SetBoxProfileSource(systemA, &systemB)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = systemModel.SetBoxProfileSource(systemB, &systemC)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = systemModel.SetBoxProfileSource(systemC, &systemA)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "would create a cycle")
	assert.False(t, ok)
}

func TestRenameGlobalDefaultCreatesOwnedDefault(t *testing.T) {
	t.Parallel()

	dbConn := setupTestDB(t)
	if dbConn == nil {
		return
	}
	defer dbConn.Close()

	userID := createTestUserID()
	systemModel := &model.BoxSystem{
		Db:     dbConn,
		UserID: userID,
	}
	defer cleanupBoxSystemsByUser(t, systemModel)
	requireLangstrothSeeded(t, systemModel)

	var globalDefaultID string
	err := dbConn.Get(&globalDefaultID, `
		SELECT CAST(id AS CHAR)
		FROM box_systems
		WHERE user_id IS NULL
		  AND is_default = 1
		  AND active = 1
		ORDER BY id ASC
		LIMIT 1
	`)
	require.NoError(t, err)
	require.NotEmpty(t, globalDefaultID)

	renamed, err := systemModel.Rename(globalDefaultID, "My Hive System")
	require.NoError(t, err)
	require.NotNil(t, renamed)
	assert.Equal(t, userID, renamed.UserID)
	assert.Equal(t, "My Hive System", renamed.Name)
	assert.True(t, renamed.IsDefault)

	var globalName string
	err = dbConn.Get(&globalName, `
		SELECT name
		FROM box_systems
		WHERE id = ?
		  AND user_id IS NULL
		  AND active = 1
		LIMIT 1
	`, globalDefaultID)
	require.NoError(t, err)
	assert.Equal(t, "Langstroth", globalName)

	visible, err := systemModel.ListVisible()
	require.NoError(t, err)
	require.NotEmpty(t, visible)

	hasOwnedDefault := false
	hasGlobalDefault := false
	for _, item := range visible {
		if item.IsDefault && item.UserID == userID {
			hasOwnedDefault = true
		}
		if item.IsDefault && item.UserID == "" {
			hasGlobalDefault = true
		}
	}
	assert.True(t, hasOwnedDefault)
	assert.False(t, hasGlobalDefault)
}
