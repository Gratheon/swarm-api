package model

import (
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type Inspection struct {
	Db     *sqlx.DB
	ID     string `json:"id" db:"id"`
	UserID string `db:"user_id"`
	HiveID string `json:"hiveId" db:"hive_id"`
	Data   string `json:"data" db:"data"`
	Added  string `json:"added" db:"added"`
}

func (r *Inspection) Get(ID string) (*Inspection, error) {
	currentInspection := Inspection{}
	err := r.Db.Get(&currentInspection,
		`SELECT *
		FROM inspections
		WHERE id=$1 AND user_id=$2
		LIMIT 1`, ID, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &currentInspection, err
}

func (r *Inspection) GetLatestByHiveId(hiveID string) (*Inspection, error) {
	currentInspection := Inspection{}
	err := r.Db.Get(&currentInspection,
		`SELECT *
			FROM inspections
			WHERE hive_id=$1 AND user_id=$2
			ORDER BY added DESC, id DESC
			LIMIT 1`, hiveID, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &currentInspection, err
}

func (r *Inspection) ListByHiveId(hiveID string) ([]*Inspection, error) {
	list := []*Inspection{}
	err := r.Db.Select(&list,
		`SELECT *
		FROM inspections
		WHERE user_id=$1 AND hive_id=$2
		ORDER BY added DESC`, r.UserID, hiveID)

	return list, err
}

func (r *Inspection) CountByHiveId(hiveID string) (int, error) {
	type InspectionCount struct {
		Count *int `db:"count"`
	}

	result := []*InspectionCount{}
	err := r.Db.Select(&result,
		`SELECT COUNT(*) as count
		FROM inspections
		WHERE user_id=$1 AND hive_id=$2
		LIMIT 1`, r.UserID, hiveID)

	if len(result) == 0 {
		return 0, nil
	}

	return *result[0].Count, err
}

func (r *Inspection) Create(data string, hiveID int) (*string, error) {
	id, err := namedInsertReturningID(r.Db,
		`INSERT INTO inspections (user_id, hive_id, data, added)
		VALUES (:userID, :hiveID, :data, NOW())
		RETURNING id`,
		map[string]interface{}{
			"userID": r.UserID,
			"data":   data,
			"hiveID": hiveID,
		},
	)

	if err != nil {
		return nil, err
	}

	strId := strconv.Itoa(int(id))
	return &strId, nil
}
