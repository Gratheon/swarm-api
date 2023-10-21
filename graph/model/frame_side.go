package model

import (
	"github.com/Gratheon/swarm-api/logger"
	"github.com/jmoiron/sqlx"
)

type FrameSide struct {
	Db                 *sqlx.DB
	ID                 *string `json:"id" db:"id"`
	UserID             string  `db:"user_id"`
}

func (FrameSide) IsEntity() {}

func (r *FrameSide) Get(id *int) (*FrameSide, error) {
	frameSide := FrameSide{}
	err := r.Db.Get(&frameSide, "SELECT * FROM `frames_sides` WHERE id=? AND user_id=? LIMIT 1", id, r.UserID)

	if err != nil {
		logger.LogError(err)
		return nil, nil
	}

	return &frameSide, nil
}

func (r *FrameSide) CreateSide(frame *FrameSide) (*int64, error) {
	result, err := r.Db.NamedExec(
		`INSERT INTO frames_sides (
		  user_id
		) VALUES (
		    :userID
		)`,
		map[string]interface{}{
			"userID":         frame.UserID,
		},
	)

	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	id, err := result.LastInsertId()

	return &id, err
}