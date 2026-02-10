package model

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type ApiaryObstacle struct {
	Db       *sqlx.DB
	ID       string   `json:"id" db:"id"`
	UserID   string   `db:"user_id"`
	ApiaryID string   `json:"apiary_id" db:"apiary_id"`
	Type     string   `json:"type" db:"type"`
	X        float64  `json:"x" db:"x"`
	Y        float64  `json:"y" db:"y"`
	Width    *float64 `json:"width" db:"width"`
	Height   *float64 `json:"height" db:"height"`
	Radius   *float64 `json:"radius" db:"radius"`
	Rotation float64  `json:"rotation" db:"rotation"`
	Label    *string  `json:"label" db:"label"`
}

func (r *ApiaryObstacle) ListByApiary(apiaryID string) ([]*ApiaryObstacle, error) {
	obstacles := []*ApiaryObstacle{}
	err := r.Db.Select(&obstacles,
		`SELECT id, user_id, apiary_id, type, x, y, width, height, radius, rotation, label
		FROM apiary_obstacles 
		WHERE apiary_id=? AND user_id=?`, apiaryID, r.UserID)
	return obstacles, err
}

func (r *ApiaryObstacle) Create(apiaryID string, obstacleType string, x float64, y float64, width *float64, height *float64, radius *float64, rotation *float64, label *string) (*ApiaryObstacle, error) {
	rot := 0.0
	if rotation != nil {
		rot = *rotation
	}

	result, err := r.Db.Exec(
		`INSERT INTO apiary_obstacles (user_id, apiary_id, type, x, y, width, height, radius, rotation, label) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.UserID, apiaryID, obstacleType, x, y, width, height, radius, rot, label)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	obstacle := &ApiaryObstacle{}
	err = r.Db.Get(obstacle,
		`SELECT id, user_id, apiary_id, type, x, y, width, height, radius, rotation, label
		FROM apiary_obstacles WHERE id=?`, id)
	return obstacle, err
}

func (r *ApiaryObstacle) Update(id string, obstacleType string, x float64, y float64, width *float64, height *float64, radius *float64, rotation *float64, label *string) (*ApiaryObstacle, error) {
	rot := 0.0
	if rotation != nil {
		rot = *rotation
	}

	_, err := r.Db.Exec(
		`UPDATE apiary_obstacles SET type=?, x=?, y=?, width=?, height=?, radius=?, rotation=?, label=? 
		WHERE id=? AND user_id=?`,
		obstacleType, x, y, width, height, radius, rot, label, id, r.UserID)
	if err != nil {
		return nil, err
	}

	obstacle := &ApiaryObstacle{}
	err = r.Db.Get(obstacle,
		`SELECT id, user_id, apiary_id, type, x, y, width, height, radius, rotation, label
		FROM apiary_obstacles WHERE id=?`, id)
	return obstacle, err
}

func (r *ApiaryObstacle) Delete(id string) (bool, error) {
	result, err := r.Db.Exec(
		`DELETE FROM apiary_obstacles WHERE id=? AND user_id=?`,
		id, r.UserID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

func (r *ApiaryObstacle) Get(id string) (*ApiaryObstacle, error) {
	obstacle := &ApiaryObstacle{}
	err := r.Db.Get(obstacle,
		`SELECT id, user_id, apiary_id, type, x, y, width, height, radius, rotation, label
		FROM apiary_obstacles 
		WHERE id=? AND user_id=?
		LIMIT 1`, id, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return obstacle, err
}
