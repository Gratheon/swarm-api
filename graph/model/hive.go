package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Hive struct {
	Db       *sqlx.DB
	ID       string `json:"id"`
	UserID   string `db:"user_id"`
	ApiaryID int    `db:"apiary_id"`
	FamilyID *int   `db:"family_id"`
	Active   *bool  `db:"active"`

	HiveNumber *int    `json:"hive_number" db:"hive_number"`
	Notes      *string `json:"notes"`
	Note       *string `json:"note"`
	Color      *string `json:"color"`
	Status     *string `json:"status"`
	Added      *string `json:"added"`
	Boxes      []*Box  `json:"boxes"`

	CollapseDate  *string `json:"collapse_date" db:"collapse_date"`
	CollapseCause *string `json:"collapse_cause" db:"collapse_cause"`

	ParentHiveID *int    `json:"parent_hive_id" db:"parent_hive_id"`
	SplitDate    *string `json:"split_date" db:"split_date"`

	MergedIntoHiveID *int    `json:"merged_into_hive_id" db:"merged_into_hive_id"`
	MergeDate        *string `json:"merge_date" db:"merge_date"`
	MergeType        *string `json:"merge_type" db:"merge_type"`
}

func (Hive) IsEntity() {}

func (r *Hive) Get(id string) (*Hive, error) {
	hive := Hive{}
	err := r.Db.Get(&hive,
		`SELECT * 
		FROM hives 
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &hive, err
}

func (r *Hive) List(userID string) ([]*Hive, error) {
	hives := []*Hive{}
	err2 := r.Db.Select(&hives,
		`SELECT * 
		FROM hives 
		WHERE user_id=? AND active=1`, userID)
	return hives, err2
}

func (r *Hive) ListByApiary(apiaryId int) ([]*Hive, error) {
	hives := []*Hive{}
	err2 := r.Db.Select(&hives,
		`SELECT * 
		FROM hives 
		WHERE apiary_id=? AND user_id=? AND active=1`, apiaryId, r.UserID)
	return hives, err2
}

func (r *Hive) Create(input HiveInput, familyID *int) (*Hive, error) {
	tx := r.Db.MustBegin()

	hiveNumber := input.HiveNumber
	if hiveNumber == nil {
		var maxNumber sql.NullInt64
		err := tx.Get(&maxNumber,
			"SELECT MAX(hive_number) FROM hives WHERE user_id=? AND active=1",
			r.UserID)
		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			return nil, err
		}
		nextNumber := 1
		if maxNumber.Valid {
			nextNumber = int(maxNumber.Int64) + 1
		}
		hiveNumber = &nextNumber
	}

	result, err := tx.NamedExec(
		"INSERT INTO hives (apiary_id, user_id, family_id, hive_number) VALUES (:apiaryID, :userID, :familyID, :hiveNumber)",
		map[string]interface{}{
			"apiaryID":   input.ApiaryID,
			"userID":     r.UserID,
			"familyID":   familyID,
			"hiveNumber": hiveNumber,
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

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	hive := Hive{}
	err = r.Db.Get(&hive, "SELECT * FROM `hives` WHERE id=? LIMIT 1", id)

	return &hive, err
}

func (r *Hive) Update(id string, notes *string, hiveNumber *int, FamilyID *string) error {
	tx := r.Db.MustBegin()

	if hiveNumber != nil {
		var existingHiveID sql.NullString
		err := tx.Get(&existingHiveID,
			"SELECT id FROM hives WHERE hive_number=? AND user_id=? AND id!=? AND active=1 LIMIT 1",
			hiveNumber, r.UserID, id)

		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			return err
		}

		if existingHiveID.Valid {
			tx.Rollback()
			return errors.New("hive number already in use by another hive")
		}
	}

	_, err := tx.NamedExec(
		"UPDATE hives SET notes=:notes, hive_number=:hiveNumber, family_id = :familyID WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":         id,
			"notes":      notes,
			"hiveNumber": hiveNumber,
			"userID":     r.UserID,
			"familyID":   FamilyID,
		},
	)
	tx.Commit()
	return err
}

func (r *Hive) Deactivate(id string) (*bool, error) {
	success := true
	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		"UPDATE hives SET active = 0 WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":     id,
			"userID": r.UserID,
		},
	)

	if err != nil {
		success = false
		return &success, err
	}

	err = tx.Commit()

	if err != nil {
		success = false
	}

	return &success, err
}

func (r *Hive) MarkAsCollapsed(id string, collapseDate time.Time, collapseCause string) error {
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		`UPDATE hives SET status='collapsed',
			collapse_date = :collapseDate, 
			collapse_cause = :collapseCause
		WHERE id=:id AND user_id=:userID`,
		map[string]interface{}{
			"id":            id,
			"userID":        r.UserID,
			"collapseDate":  collapseDate,
			"collapseCause": collapseCause,
		},
	)

	if err != nil {
		return err
	}

	err = tx.Commit()

	return err
}

func (r *Hive) GetParentHive(parentHiveID *int) (*Hive, error) {
	if parentHiveID == nil {
		return nil, nil
	}

	hive := Hive{}
	err := r.Db.Get(&hive,
		`SELECT * FROM hives WHERE id=? AND user_id=? AND active=1 LIMIT 1`,
		*parentHiveID, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &hive, err
}

func (r *Hive) GetChildHives(hiveID string) ([]*Hive, error) {
	hives := []*Hive{}
	err := r.Db.Select(&hives,
		`SELECT * FROM hives WHERE parent_hive_id=? AND user_id=? AND active=1`,
		hiveID, r.UserID)
	return hives, err
}

func (r *Hive) Split(sourceHiveID string, name string, apiaryID int, familyID *int) (*Hive, error) {
	tx := r.Db.MustBegin()

	now := time.Now()
	result, err := tx.NamedExec(
		`INSERT INTO hives (apiary_id, name, user_id, family_id, parent_hive_id, split_date, added) 
		VALUES (:apiaryID, :name, :userID, :familyID, :parentHiveID, :splitDate, :added)`,
		map[string]interface{}{
			"apiaryID":     apiaryID,
			"name":         name,
			"userID":       r.UserID,
			"familyID":     familyID,
			"parentHiveID": sourceHiveID,
			"splitDate":    now,
			"added":        now,
		},
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	hive := Hive{}
	err = r.Db.Get(&hive, "SELECT * FROM `hives` WHERE id=? LIMIT 1", id)

	return &hive, err
}

func (r *Hive) MarkAsMerged(sourceHiveID string, targetHiveID string, mergeDate time.Time, mergeType string) error {
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		`UPDATE hives SET status='merged',
			merged_into_hive_id = :targetHiveID, 
			merge_date = :mergeDate,
			merge_type = :mergeType
		WHERE id=:id AND user_id=:userID`,
		map[string]interface{}{
			"id":           sourceHiveID,
			"userID":       r.UserID,
			"targetHiveID": targetHiveID,
			"mergeDate":    mergeDate,
			"mergeType":    mergeType,
		},
	)

	if err != nil {
		return err
	}

	err = tx.Commit()

	return err
}

func (r *Hive) GetMergedIntoHive(mergedIntoHiveID *int) (*Hive, error) {
	if mergedIntoHiveID == nil {
		return nil, nil
	}

	hive := Hive{}
	err := r.Db.Get(&hive,
		`SELECT * FROM hives WHERE id=? AND user_id=? AND active=1 LIMIT 1`,
		*mergedIntoHiveID, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &hive, err
}

func (r *Hive) GetMergedFromHives(hiveID string) ([]*Hive, error) {
	hives := []*Hive{}
	err := r.Db.Select(&hives,
		`SELECT * FROM hives WHERE merged_into_hive_id=? AND user_id=? AND active=1`,
		hiveID, r.UserID)
	return hives, err
}
