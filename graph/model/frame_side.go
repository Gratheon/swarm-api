package model

import (
	"strconv"

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
		  user_id,
		  pollen,
		  honey,
		  eggs,
		  capped_brood,
		  brood,
		  queen_detected
		) VALUES (
		    :userID,
		  	:pollen,
		  	:honey,
		  	:eggs,
		  	:capped_brood,
		  	:brood,
		  	:queen_detected
		)`,
		map[string]interface{}{
			"userID":         frame.UserID,
			"pollen":         frame.PollenPercent,
			"honey":          frame.HoneyPercent,
			"eggs":           frame.EggsPercent,
			"capped_brood":   frame.CappedBroodPercent,
			"brood":          frame.BroodPercent,
			"queen_detected": frame.QueenDetected,
		},
	)

	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	id, err := result.LastInsertId()

	return &id, err
}

func (r *FrameSide) UpdateSide(frame FrameSideInput) (bool, error) {
	ok := false

	id, err := strconv.Atoi(frame.ID)
	if err != nil {
		return ok, err
	}

	exFrameSide, err := r.Get(&id)

	if err != nil {
		return ok, err
	}

	exFrameSide.BroodPercent = frame.BroodPercent
	exFrameSide.CappedBroodPercent = frame.CappedBroodPercent
	exFrameSide.EggsPercent = frame.EggsPercent
	exFrameSide.PollenPercent = frame.PollenPercent
	exFrameSide.HoneyPercent = frame.HoneyPercent

	_, err = r.Db.NamedExec(
		`UPDATE frames_sides SET
		  pollen = :pollen,
		  honey = :honey,
		  eggs = :eggs,
		  capped_brood = :capped_brood,
		  brood = :brood
		WHERE id = :id AND user_id=:userID`,
		map[string]interface{}{
			"pollen":       exFrameSide.PollenPercent,
			"honey":        exFrameSide.HoneyPercent,
			"eggs":         exFrameSide.EggsPercent,
			"capped_brood": exFrameSide.CappedBroodPercent,
			"brood":        exFrameSide.BroodPercent,
			"id":           frame.ID,
			"userID":       r.UserID,
		},
	)

	if err != nil {
		logger.LogError(err)
		return ok, err
	}

	ok = true
	return ok, err
}

func (r *FrameSide) UpdateQueenState(frame FrameSideInput) (bool, error) {
	ok := false

	id, err := strconv.Atoi(frame.ID)
	if err != nil {
		return ok, err
	}

	exFrameSide, err := r.Get(&id)

	if err != nil {
		return ok, err
	}

	exFrameSide.QueenDetected = frame.QueenDetected

	_, err = r.Db.NamedExec(
		`UPDATE frames_sides SET
          queen_detected = :queen_detected
		WHERE id = :id AND user_id=:userID`,
		map[string]interface{}{
			"queen_detected": exFrameSide.QueenDetected,
			"id":             frame.ID,
			"userID":         r.UserID,
		},
	)

	if err != nil {
		logger.LogError(err)
		return ok, err
	}

	ok = true
	return ok, err
}
