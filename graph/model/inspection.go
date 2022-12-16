package model

import (
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
)

type Inspection struct {
	Db     *sqlx.DB
	ID     string `json:"id" db:"id"`
	UserID string `db:"user_id"`
	HiveID string `db:"hive_id"`
	Data   string `json:"data" db:"data"`
	Added  string `json:"added" db:"added"`
}

func (r *Inspection) SetUp() {
	var schema = strings.Replace(`CREATE TABLE IF NOT EXISTS 'inspections' (
'id' int unsigned NOT NULL AUTO_INCREMENT,
'hive_id' int DEFAULT NULL,
'user_id' int unsigned NOT NULL,
'data' JSON,
'added' datetime DEFAULT NULL,
PRIMARY KEY ('id')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
`, "'", "`", -1)

	r.Db.MustExec(schema)
}

func (r *Inspection) Get(ID string) (*Inspection, error) {
	currentInspection := Inspection{}
	err2 := r.Db.Get(&currentInspection,
		`SELECT *
		FROM inspections
		WHERE id=? AND user_id=?
		LIMIT 1`, ID, r.UserID)

	return &currentInspection, err2
}

func (r *Inspection) GetLatestByHiveId(hiveID string) (*Inspection, error) {
	currentInspection := Inspection{}
	err2 := r.Db.Get(&currentInspection,
		`SELECT *
		FROM inspections
		WHERE hive_id=? AND user_id=?
		LIMIT 1`, hiveID, r.UserID)

	return &currentInspection, err2
}

func (r *Inspection) ListByHiveId(hiveID string) ([]*Inspection, error) {
	list := []*Inspection{}
	err := r.Db.Select(&list,
		`SELECT *
		FROM inspections
		WHERE user_id=? AND hive_id=?`, r.UserID, hiveID)

	return list, err
}

func (r *Inspection) Create(data string, hiveID int) (*string, error) {
	result, err := r.Db.NamedExec(
		`INSERT INTO inspections (user_id, hive_id, data, added)
		VALUES (:userID, :hiveID, :data, NOW())`,
		map[string]interface{}{
			"userID": r.UserID,
			"data":   data,
			"hiveID": hiveID,
		},
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	strId := strconv.Itoa(int(id))
	return &strId, err
}
