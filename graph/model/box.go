package model

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/jmoiron/sqlx"
)

func (r *Box) resolveSpecForHive(tx *sqlx.Tx, hiveID string, boxType BoxType) (*boxSpecLookupRow, error) {
	var systemID sql.NullInt64
	err := tx.Get(&systemID, `
		SELECT box_system_id
		FROM hives
		WHERE id=? AND user_id=?
		LIMIT 1
	`, hiveID, r.UserID)
	if err != nil {
		return nil, err
	}

	if systemID.Valid {
		effectiveSystemID, err := resolveEffectiveBoxProfileSystemID(tx, r.UserID, int(systemID.Int64))
		if err != nil {
			return nil, err
		}

		spec, err := getBoxSpecForTypeInSystem(tx, r.UserID, effectiveSystemID, boxType)
		if err != nil {
			return nil, err
		}
		if spec != nil {
			return spec, nil
		}
	}

	return getDefaultBoxSpecForType(tx, r.UserID, boxType)
}

type Box struct {
	Db          *sqlx.DB
	ID          *string    `json:"id"`
	UserID      string     `db:"user_id"`
	HiveId      int        `json:"hive_id" db:"hive_id"`
	Position    *int       `json:"position"`
	Color       *string    `json:"color"`
	HoleCount   *int       `json:"holeCount" db:"hole_count"`
	RoofStyle   *RoofStyle `json:"roofStyle" db:"roof_style"`
	Type        BoxType    `json:"type" db:"type"`
	BoxSystemID *int       `json:"box_system_id" db:"box_system_id"`
	BoxSpecID   *int       `json:"box_spec_id" db:"box_spec_id"`
	Active      int        `db:"active"`
}

const (
	gateHoleCountMin     = 0
	gateHoleCountMax     = 16
	gateHoleCountDefault = 8
)

func normalizeGateHoleCount(holeCount *int) int {
	value := gateHoleCountDefault
	if holeCount != nil {
		value = *holeCount
	}

	if value < gateHoleCountMin {
		value = gateHoleCountMin
	}
	if value > gateHoleCountMax {
		value = gateHoleCountMax
	}

	return value
}

func normalizeRoofStyleForType(boxType BoxType, roofStyle *RoofStyle) interface{} {
	if boxType != BoxTypeRoof {
		return nil
	}
	if roofStyle != nil && (*roofStyle == RoofStyleAngular || *roofStyle == RoofStyleFlat) {
		return *roofStyle
	}
	return RoofStyleFlat
}

func (r *Box) getHiveBoxSystemID(tx *sqlx.Tx, hiveID string) (*int, error) {
	var systemID sql.NullInt64
	err := tx.Get(&systemID, `
		SELECT box_system_id
		FROM hives
		WHERE id=? AND user_id=?
		LIMIT 1
	`, hiveID, r.UserID)
	if err != nil {
		return nil, err
	}
	if !systemID.Valid {
		return nil, nil
	}
	value := int(systemID.Int64)
	return &value, nil
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

func (r *Box) CreateByHiveId(hiveId string, boxCount int, colors []*string, boxType BoxType) error {
	tx := r.Db.MustBegin()
	spec, err := r.resolveSpecForHive(tx, hiveId, boxType)
	if err != nil {
		tx.Rollback()
		return err
	}
	if spec == nil {
		tx.Rollback()
		return errors.New("no box specification found for box type")
	}
	hiveBoxSystemID, err := r.getHiveBoxSystemID(tx, hiveId)
	if err != nil {
		tx.Rollback()
		return err
	}
	var hiveBoxSystemValue interface{} = nil
	if hiveBoxSystemID != nil {
		hiveBoxSystemValue = *hiveBoxSystemID
	}

	for position := 0; position < boxCount; position++ {
		color := "#ffc848"
		if position < len(colors) && colors[position] != nil {
			color = *colors[position]
		}

		_, err2 := tx.NamedExec(
			`INSERT INTO boxes (hive_id, position, color, user_id, type, box_system_id, box_spec_id)
			 VALUES (:hiveId, :position, :color, :userID, :type, :boxSystemID, :boxSpecID)`,
			map[string]interface{}{
				"hiveId":      hiveId,
				"position":    position,
				"color":       color,
				"userID":      r.UserID,
				"type":        boxType,
				"boxSystemID": hiveBoxSystemValue,
				"boxSpecID":   spec.BoxSpecID,
			},
		)

		if err2 != nil {
			return err2
		}
	}

	return tx.Commit()
}

func (r *Box) Create(hiveId string, position int, color *string, boxType BoxType, holeCount *int) (*string, error) {
	tx := r.Db.MustBegin()
	spec, err := r.resolveSpecForHive(tx, hiveId, boxType)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if spec == nil {
		tx.Rollback()
		return nil, errors.New("no box specification found for box type")
	}
	hiveBoxSystemID, err := r.getHiveBoxSystemID(tx, hiveId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	var hiveBoxSystemValue interface{} = nil
	if hiveBoxSystemID != nil {
		hiveBoxSystemValue = *hiveBoxSystemID
	}

	var normalizedHoleCount interface{} = nil
	if boxType == BoxTypeGate {
		normalizedHoleCount = normalizeGateHoleCount(holeCount)
	}
	normalizedRoofStyle := normalizeRoofStyleForType(boxType, nil)

	result, err := tx.NamedExec(
		`INSERT INTO boxes (hive_id, position, color, hole_count, roof_style, user_id, type, box_system_id, box_spec_id)
			VALUES (:hiveId, :position, :color, :holeCount, :roofStyle, :userID, :boxType, :boxSystemID, :boxSpecID)`,
		map[string]interface{}{
			"hiveId":      hiveId,
			"position":    position,
			"color":       color,
			"holeCount":   normalizedHoleCount,
			"roofStyle":   normalizedRoofStyle,
			"userID":      r.UserID,
			"boxType":     boxType,
			"boxSystemID": hiveBoxSystemValue,
			"boxSpecID":   spec.BoxSpecID,
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
	spec, err := r.resolveSpecForHive(tx, hiveId, boxType)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	if spec == nil {
		tx.Rollback()
		return "", errors.New("no box specification found for box type")
	}
	hiveBoxSystemID, err := r.getHiveBoxSystemID(tx, hiveId)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	var hiveBoxSystemValue interface{} = nil
	if hiveBoxSystemID != nil {
		hiveBoxSystemValue = *hiveBoxSystemID
	}

	normalizedRoofStyle := normalizeRoofStyleForType(boxType, nil)

	result, err := tx.NamedExec(
		`INSERT INTO boxes (hive_id, position, color, roof_style, user_id, type, box_system_id, box_spec_id)
			VALUES (:hiveId, :position, :color, :roofStyle, :userID, :boxType, :boxSystemID, :boxSpecID)`,
		map[string]interface{}{
			"hiveId":      hiveId,
			"position":    position,
			"color":       color,
			"roofStyle":   normalizedRoofStyle,
			"userID":      r.UserID,
			"boxType":     boxType,
			"boxSystemID": hiveBoxSystemValue,
			"boxSpecID":   spec.BoxSpecID,
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
	box1, err := r.Get(box1ID)
	if err != nil {
		return nil, err
	}
	if box1 == nil {
		ok = false
		return &ok, nil
	}

	box2, err := r.Get(box2ID)
	if err != nil {
		return nil, err
	}
	if box2 == nil {
		ok = false
		return &ok, nil
	}

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

func (r *Box) UpdateHoleCount(id string, holeCount int) (bool, error) {
	tx := r.Db.MustBegin()

	normalizedHoleCount := normalizeGateHoleCount(&holeCount)

	_, err := tx.NamedExec(
		"UPDATE boxes SET hole_count = :holeCount WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":        id,
			"holeCount": normalizedHoleCount,
			"userID":    r.UserID,
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

func (r *Box) UpdateRoofStyle(id string, roofStyle RoofStyle) (bool, error) {
	tx := r.Db.MustBegin()
	normalized := normalizeRoofStyleForType(BoxTypeRoof, &roofStyle)

	_, err := tx.NamedExec(
		"UPDATE boxes SET roof_style = :roofStyle WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":        id,
			"roofStyle": normalized,
			"userID":    r.UserID,
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
