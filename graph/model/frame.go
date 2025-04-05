package model

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Frame struct {
	Db        *sqlx.DB
	ID        int        `json:"id"`
	UserID    string     `db:"user_id"`
	BoxId     int        `db:"box_id"`
	Position  int        `json:"position"`
	Type      FrameType  `json:"type" db:"type"`
	LeftID    *int       `json:"left" db:"left_id"`
	RightID   *int       `json:"right" db:"right_id"`
	LeftSide  *FrameSide `json:"left" `
	RightSide *FrameSide `json:"right"`
	Active    int        `db:"active"`
}

func (r *Frame) Get(id int64) (*Frame, error) {
	frame := Frame{}
	err := r.Db.Get(&frame,
		`SELECT * 
		FROM frames
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1`, id, r.UserID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &frame, err
}

func (r *Frame) CreateFramesForBox(boxID *string, frameCount int) error {
	for frameNr := 0; frameNr < frameCount; frameNr++ {
		leftSide := &FrameSide{
			Db:     r.Db,
			UserID: r.UserID,
		}
		rightSide := &FrameSide{
			Db:     r.Db,
			UserID: r.UserID,
		}

		leftID, err := leftSide.CreateSide(leftSide)

		if err != nil {
			return err
		}

		rightID, err := rightSide.CreateSide(rightSide)

		if err != nil {
			return err
		}

		_, err = r.Create(
			boxID,
			frameNr,
			FrameTypeEmptyComb,
			leftID,
			rightID,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Frame) Create(boxID *string, position int, frameType FrameType, leftID *int64, rightID *int64) (*int64, error) {
	result, err := r.Db.NamedExec(
		`INSERT INTO frames (box_id, position, left_id, right_id, user_id, type) 
		VALUES (:boxID, :position, :left_id, :right_id, :userID, :frameType)`,
		map[string]interface{}{
			"boxID":     boxID,
			"position":  position,
			"left_id":   leftID,
			"right_id":  rightID,
			"userID":    r.UserID,
			"frameType": frameType,
		},
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	return &id, err
}

func (r *Frame) IsFrameWithSides(frameType FrameType) bool {
	return (frameType == `FOUNDATION` || frameType == `EMPTY_COMB` || frameType == `VOID`)
}

func (r *Frame) Update(frameID string, boxID string, position int) (*int64, error) {
	_, err := r.Db.NamedExec(
		`UPDATE frames 
		SET box_id=:boxID, position=:position
		WHERE id=:id AND user_id=:userID`,
		map[string]interface{}{
			"id":       frameID,
			"boxID":    boxID,
			"position": position,
			"userID":   r.UserID,
		},
	)

	if err != nil {
		return nil, err
	}

	return nil, err
}

func (r *Frame) DeactivateFrames(boxId *string) error {
	_, err := r.Db.NamedExec(
		`UPDATE frames 
		SET active = 0
		WHERE box_id=:boxID AND user_id=:userID`,
		map[string]interface{}{
			"boxID":  boxId,
			"userID": r.UserID,
		},
	)

	return err
}

func (r *Frame) ListByBox(boxId *string) ([]*Frame, error) {
	frames := []*Frame{}
	err := r.Db.Select(&frames,
		`SELECT id, position, left_id, right_id, type
	FROM frames
	WHERE frames.box_id =? AND user_id=? AND active=1
	ORDER BY position`, boxId, r.UserID)

	return frames, err
}

func (r *Frame) Deactivate(id string) (*bool, error) {
	success := true
	tx := r.Db.MustBegin()

	_, err := tx.NamedExec(
		`UPDATE frames 
		SET active = 0
		WHERE id=:id AND user_id=:userID`,
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

// GetFrameBySideID finds the parent Frame ID for a given FrameSide ID.
func (r *Frame) GetFrameBySideID(sideID int) (*int, error) {
	var frameId int
	err := r.Db.Get(&frameId,
		`SELECT id
		FROM frames
		WHERE (left_id = ? OR right_id = ?) AND user_id = ? AND active = 1
		LIMIT 1`, sideID, sideID, r.UserID)

	if err == sql.ErrNoRows {
		// No frame found for this side ID, might be an orphan or error
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &frameId, nil
}
