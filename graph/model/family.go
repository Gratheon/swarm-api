package model

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type Family struct {
	Db          *sqlx.DB
	UserID      string        `db:"user_id"`
	ID          string        `json:"id"  db:"id"`
	Race        *string       `json:"race" db:"race"`
	Age         *int          `json:"age"`
	Added       *string       `json:"added" db:"added"`
	Color       *string       `json:"color" db:"color"`
	Inspections []*Inspection `json:"inspections"`
}

func (r *Family) GetById(id *int) (*Family, error) {
	family := Family{}
	err := r.Db.Get(&family,
		`SELECT * 
		FROM families
		WHERE id=? AND user_id=?
		LIMIT 1`, id, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if family.Added != nil {
		// Parse the birth year from the string
		birthYear, err := strconv.Atoi(*family.Added)
		if err == nil {
			currentYear := time.Now().Year()

			age := (currentYear - birthYear)
			family.Age = &age
		}
	}

	return &family, err
}

func (r *Family) Create(race *string, added *string, color *string) (*int, error) {
	result, err := r.Db.NamedExec(
		`INSERT INTO families (user_id, race, added, color) 
		VALUES (:userID, :race, :added, :color)`,
		map[string]interface{}{
			"userID": r.UserID,
			"race":   race,
			"added":  added,
			"color":  color,
		},
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	id2 := int(id)

	return &id2, err
}

func (r *Family) Update(id *string, race *string, added *string, color *string) (*int64, error) {
	_, err := r.Db.NamedExec(
		`UPDATE families 
		SET race=:race, added = :added, color = :color
		WHERE id=:id AND user_id=:userID`,
		map[string]interface{}{
			"id":     id,
			"race":   race,
			"added":  added,
			"color":  color,
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
		_, err := r.Update(FamilyID, hive.Family.Race, hive.Family.Added, hive.Family.Color)

		if err != nil {
			return nil, err
		}
	} else {
		familyIDInt, err := r.Create(hive.Family.Race, hive.Family.Added, hive.Family.Color)
		if err != nil {
			return nil, err
		}
		familyid := strconv.Itoa(*familyIDInt)
		FamilyID = &familyid
	}
	return FamilyID, nil
}
