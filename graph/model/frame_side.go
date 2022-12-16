package model

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/gratheon/swarm-api/logger"
)

type FrameSide struct {
	Db                 *sqlx.DB
	ID                 *string `json:"id" db:"id"`
	UserID             string  `db:"user_id"`
	FileID             *string `db:"file_id"`
	BroodPercent       *int    `json:"broodPercent" db:"brood"`
	CappedBroodPercent *int    `json:"cappedBroodPercent" db:"capped_brood"`
	DroneBroodPercent  *int    `json:"droneBroodPercent" db:"drone_brood"`
	PollenPercent      *int    `json:"pollenPercent" db:"pollen"`
	HoneyPercent       *int    `json:"honeyPercent" db:"honey"`
	QueenDetected      bool    `json:"queenDetected" db:"queen_detected"`
	WorkerCount      bool    `json:"workerCount" db:"workers"`
	DroneCount      bool    `json:"droneCount" db:"drones"`
}

func (FrameSide) IsEntity() {}

func (r *FrameSide) SetUp() {
	var schema = strings.Replace(`
		CREATE TABLE IF NOT EXISTS 'frames_sides' (
		  'id' int unsigned NOT NULL AUTO_INCREMENT,
		  'user_id' int unsigned NOT NULL,
		  'file_id' int unsigned DEFAULT NULL,
		  'brood' int DEFAULT NULL,
		  'capped_brood' int DEFAULT NULL,
		  'drone_brood' int DEFAULT NULL,
		  'pollen' int DEFAULT NULL,
		  'honey' int DEFAULT NULL,
		  'queen_detected' tinyint(1) NOT NULL DEFAULT 0,
		  KEY 'file_id' ('file_id'),
		  PRIMARY KEY ('id'),
		  CONSTRAINT 'frames_sides_ibfk_1' FOREIGN KEY ('file_id') REFERENCES 'files' ('id')
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
		`, "'", "`", -1)

	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	r.Db.MustExec("SET FOREIGN_KEY_CHECKS=0;")
	r.Db.MustExec(schema)
	r.Db.MustExec("SET FOREIGN_KEY_CHECKS=1;")
}

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
		  drone_brood,
		  capped_brood,
		  brood,
		  queen_detected
		) VALUES (
		    :userID,
		  	:pollen,
		  	:honey,
		  	:drone_brood,
		  	:capped_brood,
		  	:brood,
		  	:queen_detected
		)`,
		map[string]interface{}{
			"userID":         frame.UserID,
			"pollen":         frame.PollenPercent,
			"honey":          frame.HoneyPercent,
			"drone_brood":    frame.DroneBroodPercent,
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

func (r *FrameSide) UpdateSide(frame *FrameSide) (*int64, error) {
	result, err := r.Db.NamedExec(
		`UPDATE frames_sides SET
		  file_id = :fileID,
		  pollen = :pollen,
		  honey = :honey,
		  drone_brood = :drone_brood,
		  capped_brood = :capped_brood,
		  brood = :brood,
          queen_detected = :queen_detected
		WHERE id = :id AND user_id=:userID`,
		map[string]interface{}{
			"fileID":         frame.FileID,
			"pollen":         frame.PollenPercent,
			"honey":          frame.HoneyPercent,
			"drone_brood":    frame.DroneBroodPercent,
			"capped_brood":   frame.CappedBroodPercent,
			"brood":          frame.BroodPercent,
			"queen_detected": frame.QueenDetected,
			"id":             frame.ID,
			"userID":         r.UserID,
		},
	)

	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	id, err := result.LastInsertId()

	return &id, err
}