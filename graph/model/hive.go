package model

import (
	"github.com/jmoiron/sqlx"
)

type Hive struct {
	Db       *sqlx.DB
	ID       string  `json:"id"`
	UserID   string  `db:"user_id"`
	ApiaryID int     `db:"apiary_id"`
	FamilyID *int    `db:"family_id"`
	Active   *bool   `db:"active"`
	Name     *string `json:"name"`
	Notes    *string `json:"notes"`
	Note     *string `json:"note"`
	Color    *string `json:"color"`
	Boxes    []*Box  `json:"boxes"`
}

func (Hive) IsEntity() {}

func (r *Hive) Get(id string) (*Hive, error) {
	hive := Hive{}
	err := r.Db.Get(&hive,
		`SELECT * 
		FROM hives 
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

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

func (r *Hive) Create(input HiveInput) (*Hive, error) {
	tx := r.Db.MustBegin()

	result, err := tx.NamedExec(
		"INSERT INTO hives (apiary_id, name, user_id) VALUES (:apiaryID, :name, :userID)",
		map[string]interface{}{
			"apiaryID": input.ApiaryID,
			"name":     input.Name,
			"userID":   r.UserID,
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

func (r *Hive) Update(id string, name *string, notes *string, FamilyID *string) error {
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		"UPDATE hives SET name = :name, notes=:notes, family_id = :familyID WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":       id,
			"name":     name,
			"notes":     notes,
			"userID":   r.UserID,
			"familyID": FamilyID,
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
	err = tx.Commit()

	if err != nil {
		success = false
	}

	return &success, err
}
