package model

import (
	"database/sql"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Treatment struct {
	Db *sqlx.DB

	ID       int     `json:"id"`
	UserID   string  `json:"user_id" db:"user_id"`
	BoxId    *string `json:"box_id" db:"box_id"`
	HiveId   string  `json:"hive_id" db:"hive_id"`
	FamilyId string  `json:"family_id" db:"family_id"`
	Added    string  `json:"added" db:"added"`
	Type     string  `json:"type" db:"type"`
}

func (Treatment) IsEntity() {}

func (r *Treatment) Get(id string) (*Treatment, error) {
	result := Treatment{}
	err2 := r.Db.Get(&result,
		`SELECT * 
		FROM treatments 
		WHERE id=? AND user_id=?
		LIMIT 1`, id, r.UserID)

	if err2 == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err2
}

func (r *Treatment) ListFamilyTreatments(familyId string) ([]*Treatment, error) {
	results := []*Treatment{}
	err2 := r.Db.Select(&results,
		`SELECT * 
		FROM treatments
		WHERE user_id=? AND family_id=?
		ORDER BY added DESC
		LIMIT 30`, r.UserID, familyId)

	return results, err2
}

func (r *Treatment) GetLastFamilyTreatment(familyId string) (*Treatment, error) {
	result := Treatment{}
	err2 := r.Db.Get(&result,
		`SELECT * 
		FROM treatments
		WHERE user_id=? AND family_id=?
		ORDER BY added DESC
		LIMIT 1`, r.UserID, familyId)

	if err2 == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err2
}

func (r *Treatment) TreatHive(input TreatmentOfHiveInput, familyId *int) (*Treatment, error) {
	tx := r.Db.MustBegin()

	result, err := tx.NamedExec(
		"INSERT INTO treatments (type, hive_id, user_id, family_id) VALUES (:type, :hiveId, :userID, :familyId)",
		map[string]interface{}{
			"userID":   r.UserID,
			"type":     input.Type,
			"hiveId":   input.HiveID,
			"familyId": familyId,
		})

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	tx.Commit()

	if err != nil {
		return nil, err
	}

	return r.Get(strconv.Itoa(int(id)))
}

func (r *Treatment) TreatHiveBox(input TreatmentOfBoxInput, familyId *int) (*Treatment, error) {
	tx := r.Db.MustBegin()

	result, err := tx.NamedExec(
		"INSERT INTO treatments (type, hive_id, box_id, user_id, family_id) VALUES (:type, :hiveId, :boxId, :userID, familyId)",
		map[string]interface{}{
			"userID":   r.UserID,
			"type":     input.Type,
			"hiveId":   input.HiveID,
			"familyId": familyId,
			"boxId":    input.BoxID,
		})

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	tx.Commit()

	if err != nil {
		return nil, err
	}

	return r.Get(strconv.Itoa(int(id)))
}
