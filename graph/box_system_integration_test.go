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

	userID := "999301"

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

	userID := "999302"

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
