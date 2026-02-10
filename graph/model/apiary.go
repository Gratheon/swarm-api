package model

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Apiary struct {
	Db       *sqlx.DB
	ID       int     `json:"id"`
	UserID   string  `json:"user_id" db:"user_id"`
	Name     *string `json:"name"`
	Location *string `json:"location"`
	Active   *bool   `json:"active" db:"active"`
	Lat      *string `json:"lat" db:"lat"`
	Lng      *string `json:"lng" db:"lng"`
}

func (Apiary) IsEntity() {}

func (r *Apiary) Get(id string) (*Apiary, error) {
	apiary := Apiary{}
	err2 := r.Db.Get(&apiary,
		`SELECT * 
		FROM apiaries 
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

	if err2 == sql.ErrNoRows {
		return nil, nil
	}

	return &apiary, err2
}

func (r *Apiary) List() ([]*Apiary, error) {
	apiaries := []*Apiary{}
	err2 := r.Db.Select(&apiaries,
		`SELECT * 
		FROM apiaries
		WHERE user_id=? AND active=1`, r.UserID)

	return apiaries, err2
}

func (r *Apiary) Create(input ApiaryInput) (*Apiary, error) {
	tx := r.Db.MustBegin()

	result, err := tx.NamedExec(
		"INSERT INTO apiaries (name, lat, lng, user_id) VALUES (:name, :lat, :lng, :userID)",
		map[string]interface{}{
			"userID": r.UserID,
			"name":   input.Name,
			"lat":    input.Lat,
			"lng":    input.Lng,
		})

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	tx.Commit()

	if err != nil {
		return nil, err
	}

	apiary := Apiary{}
	err = r.Db.Get(&apiary, "SELECT * FROM `apiaries` WHERE id=? AND user_id=? LIMIT 1", id, r.UserID)

	return &apiary, err
}

func (r *Apiary) Update(id string, input ApiaryInput) (*Apiary, error) {
	tx := r.Db.MustBegin()

	_, err2 := tx.NamedExec(
		"UPDATE `apiaries` SET name = :name, lat = :lat, lng = :lng WHERE id=:id",
		map[string]interface{}{
			"id":   id,
			"name": input.Name,
			"lat":  input.Lat,
			"lng":  input.Lng,
		})

	tx.Commit()

	if err2 != nil {
		return nil, err2
	}

	apiary := Apiary{}
	err3 := r.Db.Get(&apiary,
		`SELECT * 
		FROM apiaries 
		WHERE id=? AND user_id=?
		LIMIT 1`, id, r.UserID)

	return &apiary, err3
}

func (r *Apiary) Deactivate(id string) (*bool, error) {
	success := true
	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		"UPDATE apiaries SET active = 0 WHERE id=:id AND user_id=:userID",
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
