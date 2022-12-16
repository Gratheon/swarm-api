package graph

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/gratheon/swarm-api/graph/model"
	"gitlab.com/gratheon/swarm-api/logger"
)

func _upsertFamily(Db *sqlx.DB, uid string, hive model.HiveUpdateInput) *string {
	familyModel := &model.Family{
		Db:     Db,
		UserID: uid,
	}

	if hive.Family == nil {
		return nil
	}

	FamilyID := hive.Family.ID
	if hive.Family.ID != nil {
		_, err := familyModel.Update(FamilyID, hive.Family.Race, hive.Family.Added)

		if err != nil {
			logger.LogError(err)
		}
	} else {
		FamilyID, _ = familyModel.Create(hive.Family.Race, hive.Family.Added)
	}
	return FamilyID
}
