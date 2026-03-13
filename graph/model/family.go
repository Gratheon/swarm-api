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

const (
	familyMoveTypeAssigned    = "ASSIGNED"
	familyMoveTypeTransferred = "TRANSFERRED"
	familyMoveTypeWarehouse   = "WAREHOUSE"
	familyMoveTypeDeleted     = "DELETED"
)

func (r *Family) createMoveTx(tx *sqlx.Tx, familyID int, fromHiveID *int, toHiveID *int, moveType string) error {
	_, err := tx.NamedExec(
		`INSERT INTO family_moves (user_id, family_id, from_hive_id, to_hive_id, move_type)
		VALUES (:userID, :familyID, :fromHiveID, :toHiveID, :moveType)`,
		map[string]interface{}{
			"userID":     r.UserID,
			"familyID":   familyID,
			"fromHiveID": fromHiveID,
			"toHiveID":   toHiveID,
			"moveType":   moveType,
		},
	)
	return err
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

	tx := r.Db.MustBegin()

	result, err := tx.NamedExec(
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
		tx.Rollback()
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	id2 := int(id)

	if err = r.createMoveTx(tx, id2, nil, &hiveIDInt, familyMoveTypeAssigned); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &id2, err
}

func (r *Family) MoveToWarehouse(hiveID string, familyID string) (*Family, error) {
	hiveIDInt, err := strconv.Atoi(hiveID)
	if err != nil {
		return nil, err
	}
	familyIDInt, err := strconv.Atoi(familyID)
	if err != nil {
		return nil, err
	}

	tx := r.Db.MustBegin()

	result, err := tx.Exec(
		`UPDATE families
		SET hive_id=NULL
		WHERE id=? AND hive_id=? AND user_id=?`,
		familyID, hiveID, r.UserID,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return nil, sql.ErrNoRows
	}

	if err = r.createMoveTx(tx, familyIDInt, &hiveIDInt, nil, familyMoveTypeWarehouse); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetById(&familyIDInt)
}

func (r *Family) MoveBetweenHives(familyID string, fromHiveID string, toHiveID string) error {
	familyIDInt, err := strconv.Atoi(familyID)
	if err != nil {
		return err
	}
	fromHiveIDInt, err := strconv.Atoi(fromHiveID)
	if err != nil {
		return err
	}
	toHiveIDInt, err := strconv.Atoi(toHiveID)
	if err != nil {
		return err
	}

	tx := r.Db.MustBegin()

	result, err := tx.Exec(
		`UPDATE families
		SET hive_id=?
		WHERE id=? AND hive_id=? AND user_id=?`,
		toHiveIDInt, familyIDInt, fromHiveIDInt, r.UserID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return sql.ErrNoRows
	}

	if err = r.createMoveTx(tx, familyIDInt, &fromHiveIDInt, &toHiveIDInt, familyMoveTypeTransferred); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *Family) AssignFromWarehouse(hiveID string, familyID string) (*Family, error) {
	hiveIDInt, err := strconv.Atoi(hiveID)
	if err != nil {
		return nil, err
	}
	familyIDInt, err := strconv.Atoi(familyID)
	if err != nil {
		return nil, err
	}

	tx := r.Db.MustBegin()

	result, err := tx.Exec(
		`UPDATE families
		SET hive_id=?
		WHERE id=? AND hive_id IS NULL AND user_id=?`,
		hiveIDInt, familyIDInt, r.UserID,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return nil, sql.ErrNoRows
	}

	if err = r.createMoveTx(tx, familyIDInt, nil, &hiveIDInt, familyMoveTypeAssigned); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetById(&familyIDInt)
}

func (r *Family) DeleteFromHive(hiveID string, familyID string) (bool, error) {
	hiveIDInt, err := strconv.Atoi(hiveID)
	if err != nil {
		return false, err
	}
	familyIDInt, err := strconv.Atoi(familyID)
	if err != nil {
		return false, err
	}

	tx := r.Db.MustBegin()

	result, err := tx.Exec(
		`DELETE FROM families
		WHERE id=? AND hive_id=? AND user_id=?`,
		familyIDInt, hiveIDInt, r.UserID,
	)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return false, nil
	}

	if err = r.createMoveTx(tx, familyIDInt, &hiveIDInt, nil, familyMoveTypeDeleted); err != nil {
		tx.Rollback()
		return false, err
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Family) ListUnassigned() ([]*Family, error) {
	families := []*Family{}
	err := r.Db.Select(&families,
		`SELECT *
		FROM families
		WHERE hive_id IS NULL AND user_id=?
		ORDER BY id DESC`,
		r.UserID,
	)
	if err != nil {
		return nil, err
	}

	for _, family := range families {
		if family.Added != nil {
			birthYear, convErr := strconv.Atoi(*family.Added)
			if convErr == nil {
				currentYear := time.Now().Year()
				age := currentYear - birthYear
				family.Age = &age
			}
		}
	}

	return families, nil
}

func (r *Family) LastHiveID(familyID string) (*int, error) {
	var lastHiveID sql.NullInt64
	err := r.Db.Get(&lastHiveID,
		`SELECT COALESCE(to_hive_id, from_hive_id) AS hive_id
		FROM family_moves
		WHERE family_id=? AND user_id=?
		  AND (to_hive_id IS NOT NULL OR from_hive_id IS NOT NULL)
		ORDER BY moved_at DESC, id DESC
		LIMIT 1`,
		familyID, r.UserID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !lastHiveID.Valid {
		return nil, nil
	}

	id := int(lastHiveID.Int64)
	return &id, nil
}
