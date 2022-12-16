package model

import (
	//"fmt"
	"strings"
	//"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Apiary struct {
	Db       *sqlx.DB
	ID       int     `json:"id"`
	UserID   string  `db:"user_id"`
	Name     *string `json:"name"`
	Location *string `json:"location"`
	Active   *bool   `db:"active"`
	Lat      *string `json:"lat" db:"lat"`
	Lng      *string `json:"lng" db:"lng"`
	//Hives    []*Hive `json:"hives"`
}

func (Apiary) IsEntity() {}

func (r *Apiary) SetUp() {
	var schema = strings.Replace(
		`CREATE TABLE IF NOT EXISTS 'apiaries' (
  'id' int unsigned NOT NULL AUTO_INCREMENT,
  'user_id' int unsigned NOT NULL,
  'name' varchar(250) DEFAULT NULL,
	'active' tinyint(1) NOT NULL DEFAULT 1,
  'lng' varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT "0",
  'lat' varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT "0",
  PRIMARY KEY ('id')
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
`, "'", "`", -1)

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	r.Db.MustExec(schema)
}

func (r *Apiary) Get(id string) (*Apiary, error) {
	apiary := Apiary{}
	err2 := r.Db.Get(&apiary,
		`SELECT * 
		FROM apiaries 
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

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
