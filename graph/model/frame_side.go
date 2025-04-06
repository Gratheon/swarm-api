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

	FrameID            *int    `json:"frameId" db:"frame_id"` // Add FrameID field back for optimized query
}

func (FrameSide) IsEntity() {}

func (r *FrameSide) Get(id *int) (*FrameSide, error) {
	frameSide := FrameSide{}
	// Updated query to LEFT JOIN frames and select frame_id
	query := `
		SELECT fs.*, f.id as frame_id
		FROM frames_sides fs
		LEFT JOIN frames f ON (fs.id = f.left_id OR fs.id = f.right_id) AND f.user_id = fs.user_id AND f.active = 1
		WHERE fs.id = ? AND fs.user_id = ?
		LIMIT 1`
	err := r.Db.Get(&frameSide, query, id, r.UserID)

	if err == sql.ErrNoRows {
		// It's possible the FrameSide exists but has no parent Frame (orphan)
		// Try fetching just the FrameSide data in that case
		errFallback := r.Db.Get(&frameSide, "SELECT * FROM `frames_sides` WHERE id=? AND user_id=? LIMIT 1", id, r.UserID)
		if errFallback == sql.ErrNoRows {
			return nil, nil // Truly doesn't exist
		}
		if errFallback != nil {
			logger.LogError(errFallback)
			return nil, errFallback // Error during fallback fetch
		}
		// FrameSide exists but no parent frame found, FrameID will be nil
		return &frameSide, nil
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

