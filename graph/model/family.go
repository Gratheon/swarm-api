package model

import (
	"github.com/jmoiron/sqlx"
	"strconv"
)

type Family struct {
	Db          *sqlx.DB
	UserID      string        `db:"user_id"`
	ID          string        `json:"id"  db:"id"`
	Race        *string       `json:"race" db:"race"`
	Added       *string        `json:"added" db:"added"`
	Inspections []*Inspection `json:"inspections"`
}

func (r *Family) GetById(id *int) (*Family, error) {
	family := Family{}
	err := r.Db.Get(&family,
		`SELECT * 
		FROM families
		WHERE id=? AND user_id=?
		LIMIT 1`, id, r.UserID)

	return &family, err
}

func (r *Family) Create(race *string, added *string) (*string, error) {
	result, err := r.Db.NamedExec(
		`INSERT INTO families (user_id, race, added) 
		VALUES (:userID, :race, :added)`,
		map[string]interface{}{
			"userID": r.UserID,
			"race":   race,
			"added":  added,
		},
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	strId := strconv.Itoa(int(id))
	return &strId, err
}

func (r *Family) Update(id *string, race *string, added *string) (*int64, error) {
	_, err := r.Db.NamedExec(
		`UPDATE families 
		SET race=:race, added = :added
		WHERE id=:id AND user_id=:userID`,
		map[string]interface{}{
			"id":     id,
			"race":   race,
			"added":  added,
			"userID": r.UserID,
		},
	)

	if err != nil {
		return nil, err
	}

	return nil, err
}

func (r *Family) Upsert(uid string, hive HiveUpdateInput) (*string, error) {
	if hive.Family == nil {
		return nil, nil
	}

	FamilyID := hive.Family.ID
	if hive.Family.ID != nil {
		_, err := r.Update(FamilyID, hive.Family.Race, hive.Family.Added)

		if err != nil {
			return nil, err
		}	
	} else {
		FamilyID, _ = r.Create(hive.Family.Race, hive.Family.Added)
	}
	return FamilyID, nil
}
