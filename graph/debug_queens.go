package graph

import (
	"context"
	"strconv"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/Gratheon/swarm-api/logger"
)

func (r *mutationResolver) DebugHiveQueens(ctx context.Context, hiveID string) (string, error) {
	uid := ctx.Value("userID").(string)

	logger.Info("=== DEBUG: Checking queens for hive " + hiveID + " ===")

	result := "Hive ID: " + hiveID + "\n"
	result += "Note: hives.family_id column has been removed - using only families.hive_id\n"

	familyModel := &model.Family{
		Db:     r.Resolver.Db,
		UserID: uid,
	}

	families, err := familyModel.ListByHive(hiveID)
	if err != nil {
		result += "Error listing families: " + err.Error() + "\n"
	} else {
		result += "Queens found via families.hive_id: " + strconv.Itoa(len(families)) + "\n"
		for i, family := range families {
			result += "  Queen " + strconv.Itoa(i+1) + ": ID=" + family.ID
			if family.Name != nil {
				result += ", Name=" + *family.Name
			}
			result += "\n"
		}
	}

	var directCount int
	r.Resolver.Db.Get(&directCount, "SELECT COUNT(*) FROM families WHERE hive_id=? AND user_id=?", hiveID, uid)
	result += "Direct query families.hive_id count: " + strconv.Itoa(directCount) + "\n"

	logger.Info(result)

	return result, nil
}
