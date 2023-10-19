package model

import (
	"github.com/Gratheon/swarm-api/logger"
	"github.com/jmoiron/sqlx"
)

type FrameSide struct {
	Db                 *sqlx.DB
	ID                 *string `json:"id" db:"id"`
	UserID             string  `db:"user_id"`
	BroodPercent       *int    `json:"broodPercent" db:"brood"`
	CappedBroodPercent *int    `json:"cappedBroodPercent" db:"capped_brood"`
	EggsPercent        *int    `json:"eggsPercent" db:"eggs"`
	PollenPercent      *int    `json:"pollenPercent" db:"pollen"`
	HoneyPercent       *int    `json:"honeyPercent" db:"honey"`
	QueenDetected      bool    `json:"queenDetected" db:"queen_detected"`
	WorkerCount        bool    `json:"workerCount" db:"workers"`
	DroneCount         bool    `json:"droneCount" db:"drones"`
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