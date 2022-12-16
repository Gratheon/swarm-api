package model

import (
	"strings"

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
}

func (r *Frame) SetUp() {
	var schema = strings.Replace(
		`CREATE TABLE IF NOT EXISTS 'frames' (
  'id' int unsigned NOT NULL AUTO_INCREMENT,
  'user_id' int unsigned NOT NULL,
  'box_id' int unsigned DEFAULT NULL,
  'type' enum("VOID","FOUNDATION","EMPTY_COMB","PARTITION","FEEDER") CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT "EMPTY_COMB",
  'position' int unsigned DEFAULT NULL,
  'left_id' int unsigned DEFAULT NULL,
  'right_id' int unsigned DEFAULT NULL,
  'is_deleted' int DEFAULT NULL,
  PRIMARY KEY ('id'),
  KEY 'box_id' ('box_id'),
  KEY 'left_id' ('left_id'),
  KEY 'right_id' ('right_id'),
  CONSTRAINT 'frames_ibfk_1' FOREIGN KEY ('box_id') REFERENCES 'boxes' ('id') ON DELETE CASCADE,
  CONSTRAINT 'frames_ibfk_2' FOREIGN KEY ('left_id') REFERENCES 'frames_sides' ('id') ON DELETE SET NULL,
  CONSTRAINT 'frames_ibfk_3' FOREIGN KEY ('right_id') REFERENCES 'frames_sides' ('id') ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
`, "'", "`", -1)

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	r.Db.MustExec(schema)
}

func (r *Frame) CreateFramesForBox(boxID *string, frameCount int) error {
	for frameNr := 0; frameNr < frameCount; frameNr++ {
		leftSide := &FrameSide{
			Db:                 r.Db,
			UserID:             r.UserID,
			BroodPercent:       nil,
			CappedBroodPercent: nil,
			DroneBroodPercent:  nil,
			HoneyPercent:       nil,
			PollenPercent:      nil,
		}
		rightSide := &FrameSide{
			Db:                 r.Db,
			UserID:             r.UserID,
			BroodPercent:       nil,
			CappedBroodPercent: nil,
			DroneBroodPercent:  nil,
			HoneyPercent:       nil,
			PollenPercent:      nil,
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

func (r *Frame) Update(frameID *string, boxID *string, position int) (*int64, error) {
	_, err := r.Db.NamedExec(
		`UPDATE frames 
		SET box_id=:boxID,
		    is_deleted = null,
        	position=:position
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
		SET is_deleted = 1
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
	WHERE frames.box_id =? AND user_id=? AND is_deleted IS NULL
	ORDER BY position`, boxId, r.UserID)

	return frames, err
}
