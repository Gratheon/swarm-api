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
	HiveID      *int          `json:"hive_id" db:"hive_id"`
	Name        *string       `json:"name" db:"name"`
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

func (r *Family) Create(name *string, race *string, added *string, color *string) (*int, error) {
	result, err := r.Db.NamedExec(
		`INSERT INTO families (user_id, name, race, added, color) 
		VALUES (:userID, :name, :race, :added, :color)`,
		map[string]interface{}{
			"userID": r.UserID,
			"name":   name,
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

func (r *Family) Update(id *string, name *string, race *string, added *string, color *string) (*int64, error) {
	_, err := r.Db.NamedExec(
		`UPDATE families 
		SET name=:name, race=:race, added = :added, color = :color
		WHERE id=:id AND user_id=:userID`,
		map[string]interface{}{
			"id":     id,
			"name":   name,
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
		_, err := r.Update(FamilyID, hive.Family.Name, hive.Family.Race, hive.Family.Added, hive.Family.Color)

		if err != nil {
			return nil, err
		}
	} else {
		familyIDInt, err := r.Create(hive.Family.Name, hive.Family.Race, hive.Family.Added, hive.Family.Color)
		if err != nil {
			return nil, err
		}
		familyid := strconv.Itoa(*familyIDInt)
		FamilyID = &familyid
	}
	return FamilyID, nil
}

func (r *Family) ListByHive(hiveID string) ([]*Family, error) {
	families := []*Family{}
	err := r.Db.Select(&families,
		`SELECT * 
		FROM families
		WHERE hive_id=? AND user_id=?`,
		hiveID, r.UserID)

	if err != nil {
		return nil, err
	}

	for _, family := range families {
		if family.Added != nil {
			birthYear, err := strconv.Atoi(*family.Added)
			if err == nil {
				currentYear := time.Now().Year()
				age := currentYear - birthYear
				family.Age = &age
			}
		}
	}

	return families, nil
}

func (r *Family) CreateForHive(hiveID string, name *string, race *string, added *string, color *string) (*int, error) {
	hiveIDInt, err := strconv.Atoi(hiveID)
	if err != nil {
		return nil, err
	}

	result, err := r.Db.NamedExec(
		`INSERT INTO families (user_id, hive_id, name, race, added, color) 
		VALUES (:userID, :hiveID, :name, :race, :added, :color)`,
		map[string]interface{}{
			"userID": r.UserID,
			"hiveID": hiveIDInt,
			"name":   name,
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
