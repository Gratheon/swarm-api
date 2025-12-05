package model

import (
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
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

func (r *Box) Get(id string) (*Box, error) {
	box := Box{}
	err := r.Db.Get(&box,
		`SELECT * 
		FROM boxes
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

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

func (r *Box) CreateSingleBox(hiveId string, position int, color string, boxType BoxType) (string, error) {
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
		return "", err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(id)), nil
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

func (r *Box) SwapBoxPositions(box1ID string, box2ID string) (*bool, error) {
	ok := true
	box1, _ := r.Get(box1ID)
	box2, _ := r.Get(box2ID)

	tmpPosition := *box1.Position
	box1.Position = box2.Position
	box2.Position = &tmpPosition

	_, upd1 := r.Update(box1.ID, *box1.Position, box1.Color)

	if upd1 != nil {
		ok = false
		return &ok, upd1
	}
	_, upd2 := r.Update(box2.ID, *box2.Position, box2.Color)

	if upd2 != nil {
		ok = false
		return &ok, upd2
	}

	return &ok, nil
}

func (r *Box) Update(id *string, position int, color *string) (bool, error) {
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		"UPDATE boxes SET position = :position, color = :color WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":       id,
			"position": position,
			"color":    color,
			"userID":   r.UserID,
		},
	)

	if err != nil {
		return false, err
	}

	err = tx.Commit()

	if err != nil {
		return false, err
	}

	return true, nil
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

func (r *Box) MoveBoxesToHive(boxIDs []string, targetHiveID string, startPosition int) error {
	tx := r.Db.MustBegin()

	for i, boxID := range boxIDs {
		_, err := tx.NamedExec(
			`UPDATE boxes 
			SET hive_id=:hiveID, position=:position
			WHERE id=:id AND user_id=:userID AND active=1`,
			map[string]interface{}{
				"id":       boxID,
				"hiveID":   targetHiveID,
				"position": startPosition + i,
				"userID":   r.UserID,
			},
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Box) GetMaxPosition(hiveID string) (int, error) {
	var maxPosition sql.NullInt64
	err := r.Db.QueryRow(
		`SELECT MAX(position) FROM boxes WHERE hive_id=? AND user_id=? AND active=1`,
		hiveID, r.UserID,
	).Scan(&maxPosition)

	if err != nil {
		return -1, err
	}

	if !maxPosition.Valid {
		return -1, nil
	}

	return int(maxPosition.Int64), nil
}

func (r *Box) GetBoxesByTypeForHive(hiveID string, boxTypes []BoxType) ([]*Box, error) {
	if len(boxTypes) == 0 {
		return []*Box{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT * FROM boxes WHERE hive_id=? AND user_id=? AND active=1 AND type IN (?)`,
		hiveID, r.UserID, boxTypes,
	)
	if err != nil {
		return nil, err
	}

	query = r.Db.Rebind(query)
	boxes := []*Box{}
	err = r.Db.Select(&boxes, query, args...)
	return boxes, err
}
