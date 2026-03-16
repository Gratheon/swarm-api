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

func (r *WarehouseModule) List() ([]*WarehouseModule, error) {
	rows := []*WarehouseModule{}
	err := r.Db.Select(&rows,
		`SELECT user_id, module_type, box_system_id, count
		FROM warehouse_modules
		WHERE user_id=?
		  AND box_system_id = 0
		ORDER BY module_type ASC`,
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
		ON DUPLICATE KEY UPDATE count=:count`,
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
		WHERE user_id=? AND module_type=? AND box_system_id=?
		LIMIT 1`,
		r.UserID, moduleType, systemID,
	)
	if err != nil {
		return nil, err
	}

	return &current, nil
}
