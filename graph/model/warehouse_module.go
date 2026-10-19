package model

import (
	"strings"

	"github.com/jmoiron/sqlx"
)

type WarehouseModule struct {
	Db *sqlx.DB

	UserID      string              `json:"user_id" db:"user_id"`
	ModuleType  WarehouseModuleType `json:"moduleType" db:"module_type"`
	BoxSystemID int                 `json:"box_system_id" db:"box_system_id"`
	Count       int                 `json:"count" db:"count"`
}

// List keeps the order of the former MySQL ENUM definition (ENUM columns sorted by declaration index).
func (r *WarehouseModule) List() ([]*WarehouseModule, error) {
	rows := []*WarehouseModule{}
	err := r.Db.Select(&rows,
		`SELECT user_id, module_type, box_system_id, count
		FROM warehouse_modules
		WHERE user_id=$1
		  AND box_system_id = 0
		ORDER BY array_position(
			ARRAY['DEEP', 'NUCS', 'SUPER', 'LARGE_HORIZONTAL_SECTION', 'ROOF', 'HORIZONTAL_FEEDER', 'QUEEN_EXCLUDER', 'BOTTOM',
			      'FRAME_FOUNDATION', 'FRAME_EMPTY_COMB', 'FRAME_PARTITION', 'FRAME_FEEDER']::varchar[],
			module_type
		)`,
		r.UserID,
	)
	return rows, err
}

func (r *WarehouseModule) Upsert(moduleType WarehouseModuleType, count int) (*WarehouseModule, error) {
	return r.UpsertForSystem(moduleType, nil, count)
}

func (r *WarehouseModule) UpsertForSystem(moduleType WarehouseModuleType, boxSystemID *int, count int) (*WarehouseModule, error) {
	if count < 0 {
		count = 0
	}
	systemID := 0
	if boxSystemID != nil && *boxSystemID > 0 {
		systemID = *boxSystemID
	}

	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		`INSERT INTO warehouse_modules (user_id, module_type, box_system_id, count)
		VALUES (:userID, :moduleType, :boxSystemID, :count)
		ON CONFLICT (user_id, module_type, box_system_id) DO UPDATE SET count = EXCLUDED.count`,
		map[string]interface{}{
			"userID":      r.UserID,
			"moduleType":  strings.TrimSpace(moduleType.String()),
			"boxSystemID": systemID,
			"count":       count,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	current := WarehouseModule{}
	err = r.Db.Get(&current,
		`SELECT user_id, module_type, box_system_id, count
		FROM warehouse_modules
		WHERE user_id=$1 AND module_type=$2 AND box_system_id=$3
		LIMIT 1`,
		r.UserID, moduleType, systemID,
	)
	if err != nil {
		return nil, err
	}

	return &current, nil
}
