package model

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type HivePlacement struct {
	Db       *sqlx.DB
	ID       string  `json:"id" db:"id"`
	UserID   string  `db:"user_id"`
	ApiaryID string  `json:"apiary_id" db:"apiary_id"`
	HiveID   string  `json:"hive_id" db:"hive_id"`
	X        float64 `json:"x" db:"x"`
	Y        float64 `json:"y" db:"y"`
	Rotation float64 `json:"rotation" db:"rotation"`
}

func (r *HivePlacement) ListByApiary(apiaryID string) ([]*HivePlacement, error) {
	placements := []*HivePlacement{}
	err := r.Db.Select(&placements,
		`SELECT hp.id, hp.user_id, hp.apiary_id, hp.hive_id, hp.x, hp.y, hp.rotation
		FROM hive_placements hp
		INNER JOIN hives h ON hp.hive_id = h.id AND hp.user_id = h.user_id
		WHERE hp.apiary_id=? AND hp.user_id=? 
		  AND h.active=1 
		  AND h.collapse_date IS NULL 
		  AND h.merged_into_hive_id IS NULL`, apiaryID, r.UserID)
	return placements, err
}

func (r *HivePlacement) Update(apiaryID string, hiveID string, x float64, y float64, rotation float64) (*HivePlacement, error) {
	var existingID int64
	err := r.Db.Get(&existingID,
		`SELECT id FROM hive_placements WHERE apiary_id=? AND hive_id=? AND user_id=?`,
		apiaryID, hiveID, r.UserID)

	if err == sql.ErrNoRows {
		result, err := r.Db.Exec(
			`INSERT INTO hive_placements (user_id, apiary_id, hive_id, x, y, rotation) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			r.UserID, apiaryID, hiveID, x, y, rotation)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		existingID = id
	} else if err != nil {
		return nil, err
	} else {
		_, err = r.Db.Exec(
			`UPDATE hive_placements SET x=?, y=?, rotation=? WHERE id=? AND user_id=?`,
			x, y, rotation, existingID, r.UserID)
		if err != nil {
			return nil, err
		}
	}

	placement := &HivePlacement{}
	err = r.Db.Get(placement,
		`SELECT id, user_id, apiary_id, hive_id, x, y, rotation
		FROM hive_placements WHERE id=?`, existingID)
	return placement, err
}

func (r *HivePlacement) DeleteByHiveID(hiveID string) error {
	_, err := r.Db.Exec(
		`DELETE FROM hive_placements WHERE hive_id=? AND user_id=?`,
		hiveID, r.UserID)
	return err
}
