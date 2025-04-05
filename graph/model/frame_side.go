package model

import (
	"database/sql"

	"github.com/Gratheon/swarm-api/logger"
	"github.com/jmoiron/sqlx"
)

type FrameSide struct {
	Db                 *sqlx.DB
	ID                 *string `json:"id" db:"id"`
	UserID             string  `db:"user_id"`
	IsQueenConfirmed   bool    `json:"isQueenConfirmed" db:"is_queen_confirmed"`
	// FrameID            *int    `json:"frameId" db:"frame_id"` // Reverted - Field does not exist in DB table
}

func (FrameSide) IsEntity() {}

func (r *FrameSide) Get(id *int) (*FrameSide, error) {
	frameSide := FrameSide{}
	err := r.Db.Get(&frameSide, "SELECT * FROM `frames_sides` WHERE id=? AND user_id=? LIMIT 1", id, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	
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

// UpdateQueenConfirmation updates the is_queen_confirmed status for a specific frame side.
func (r *FrameSide) UpdateQueenConfirmation(id int, confirmed bool) error {
	_, err := r.Db.Exec(
		"UPDATE frames_sides SET is_queen_confirmed = ? WHERE id = ? AND user_id = ?",
		confirmed,
		id,
		r.UserID,
	)

	if err != nil {
		logger.LogError(err)
		return err
	}

	return nil
}

// Reverted - This method is no longer needed as frame_id is not on frames_sides
// // UpdateFrameId sets the frame_id for a specific frame side.
// func (r *FrameSide) UpdateFrameId(sideId int64, frameId int64) error {
// 	_, err := r.Db.Exec(
// 		"UPDATE frames_sides SET frame_id = ? WHERE id = ? AND user_id = ?",
// 		frameId,
// 		sideId,
// 		r.UserID,
// 	)

// 	if err != nil {
// 		logger.LogError(err)
// 		return err
// 	}

// 	return nil
// }
