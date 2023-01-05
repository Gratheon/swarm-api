package model

import (
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
)

type Box struct {
	Db       *sqlx.DB
	ID       *string `json:"id"`
	UserID   string  `db:"user_id"`
	HiveId   int     `json:"hive_id" db:"hive_id"`
	Position *int    `json:"position"`
	Color    *string `json:"color"`
	Type     BoxType `json:"type" db:"type"`
	Active   int     `db:"active"`
}

func (r *Box) SetUp() {
	var schema = strings.Replace(
		`CREATE TABLE IF NOT EXISTS 'boxes' (
  'id' int unsigned NOT NULL AUTO_INCREMENT,
  'user_id' int unsigned NOT NULL,
  'hive_id' int NOT NULL,
  'active' tinyint(1) NOT NULL DEFAULT 1,
  'color' varchar(10) DEFAULT NULL,
  'position' mediumint DEFAULT NULL,
  'type' enum("SUPER","DEEP") COLLATE utf8mb4_general_ci NOT NULL DEFAULT "DEEP",
  PRIMARY KEY ('id')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
`, "'", "`", -1)

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	r.Db.MustExec(schema)
}


func (r *Box) Get(id string) (*Box, error) {
	box := Box{}
	err := r.Db.Get(&box,
		`SELECT * 
		FROM boxes
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

	return &box, err
}

func (r *Box) CreateByHiveId(hiveId string, boxCount int, colors []*string) error {
	tx := r.Db.MustBegin()

	for position := 0; position < boxCount; position++ {
		_, err2 := tx.NamedExec(
			"INSERT INTO boxes (hive_id, position, color, user_id) VALUES (:hiveId, :position, :color, :userID)",
			map[string]interface{}{
				"hiveId":   hiveId,
				"position": position,
				"color":    colors[position],
				"userID":   r.UserID,
			},
		)

		if err2 != nil {
			return err2
		}
	}

	return tx.Commit()
}

func (r *Box) Create(hiveId string, position int, color *string, boxType BoxType) (*string, error) {
	tx := r.Db.MustBegin()

	result, err := tx.NamedExec(
		`INSERT INTO boxes (hive_id, position, color, user_id, type)
			VALUES (:hiveId, :position, :color, :userID, :boxType)`,
		map[string]interface{}{
			"hiveId":   hiveId,
			"position": position,
			"color":    color,
			"userID":   r.UserID,
			"boxType":  boxType,
		},
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	strId := strconv.Itoa(int(id))

	if err != nil {
		return nil, err
	}

	return &strId, tx.Commit()
}

func (r *Box) ListByHive(hiveId string) ([]*Box, error) {
	boxes := []*Box{}
	err2 := r.Db.Select(&boxes,
		`SELECT *
		FROM boxes
		WHERE active=1 AND hive_id=? AND user_id=?
		ORDER BY position DESC`, hiveId, r.UserID)
	return boxes, err2
}

func (r *Box) Count(hiveId string) (int, error) {
	row := r.Db.QueryRow(
		`SELECT COUNT(id) as cnt
		FROM boxes
		WHERE active=1 AND hive_id=? AND user_id=?`, hiveId, r.UserID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}

func (r *Box) Update(id *string, position int, color *string, active int) error {
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		"UPDATE boxes SET position = :position, color = :color, active = :active WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":       id,
			"position": position,
			"color":    color,
			"active":   active,
			"userID":   r.UserID,
		},
	)

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Box) DeactivateByHive(id string) error {
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		"UPDATE boxes SET active=0 WHERE hive_id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":     id,
			"userID": r.UserID,
		},
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}


func (r *Box) Deactivate(id string) (*bool, error) {
	success := true
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		"UPDATE boxes SET active=0 WHERE id=:id AND user_id=:userID",
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