package model

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Apiary struct {
	Db       *sqlx.DB
	ID       int        `json:"id"`
	UserID   string     `json:"user_id" db:"user_id"`
	Name     *string    `json:"name"`
	Type     ApiaryType `json:"type" db:"type"`
	Location *string    `json:"location"`
	Active   *bool      `json:"active" db:"active"`
	Lat      *string    `json:"lat" db:"lat"`
	Lng      *string    `json:"lng" db:"lng"`
}

func (Apiary) IsEntity() {}

func ensureApiaryType(apiary *Apiary) {
	if apiary != nil && apiary.Type == "" {
		apiary.Type = ApiaryTypeStatic
	}
}

func (r *Apiary) Get(id string) (*Apiary, error) {
	apiary := Apiary{}
	err2 := r.Db.Get(&apiary,
		`SELECT * 
		FROM apiaries 
		WHERE id=$1 AND user_id=$2 AND active=TRUE
		LIMIT 1`, id, r.UserID)

	if err2 == sql.ErrNoRows {
		return nil, nil
	}
	ensureApiaryType(&apiary)

	return &apiary, err2
}

func (r *Apiary) List() ([]*Apiary, error) {
	apiaries := []*Apiary{}
	err2 := r.Db.Select(&apiaries,
		`SELECT * 
		FROM apiaries
		WHERE user_id=$1 AND active=TRUE`, r.UserID)
	for _, apiary := range apiaries {
		ensureApiaryType(apiary)
	}

	return apiaries, err2
}

func (r *Apiary) Create(input ApiaryInput) (*Apiary, error) {
	tx := r.Db.MustBegin()
	var apiaryType *string
	if input.Type != nil {
		castedType := string(*input.Type)
		apiaryType = &castedType
	}

	id, err := namedInsertReturningID(tx,
		"INSERT INTO apiaries (name, type, lat, lng, user_id) VALUES (:name, COALESCE(:type, 'STATIC'), :lat, :lng, :userID) RETURNING id",
		map[string]interface{}{
			"userID": r.UserID,
			"name":   input.Name,
			"type":   apiaryType,
			"lat":    input.Lat,
			"lng":    input.Lng,
		})

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	apiary := Apiary{}
	err = r.Db.Get(&apiary, "SELECT * FROM apiaries WHERE id=$1 AND user_id=$2 LIMIT 1", id, r.UserID)
	ensureApiaryType(&apiary)

	return &apiary, err
}

func (r *Apiary) Update(id string, input ApiaryInput) (*Apiary, error) {
	tx := r.Db.MustBegin()
	var apiaryType *string
	if input.Type != nil {
		castedType := string(*input.Type)
		apiaryType = &castedType
	}

	_, err2 := tx.NamedExec(
		"UPDATE apiaries SET name = :name, type = COALESCE(:type, type), lat = :lat, lng = :lng WHERE id=:id AND user_id = :userID",
		map[string]interface{}{
			"id":     id,
			"userID": r.UserID,
			"name":   input.Name,
			"type":   apiaryType,
			"lat":    input.Lat,
			"lng":    input.Lng,
		})

	tx.Commit()

	if err2 != nil {
		return nil, err2
	}

	apiary := Apiary{}
	err3 := r.Db.Get(&apiary,
		`SELECT * 
		FROM apiaries 
		WHERE id=$1 AND user_id=$2
		LIMIT 1`, id, r.UserID)
	ensureApiaryType(&apiary)

	return &apiary, err3
}

func (r *Apiary) Deactivate(id string) (*bool, error) {
	success := true
	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		"UPDATE apiaries SET active = FALSE WHERE id=:id AND user_id=:userID",
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
