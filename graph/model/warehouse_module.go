package model

import (
	"strings"

	"github.com/jmoiron/sqlx"
)

type WarehouseModule struct {
	Db *sqlx.DB

	UserID     string              `json:"user_id" db:"user_id"`
	ModuleType WarehouseModuleType `json:"moduleType" db:"module_type"`
	Count      int                 `json:"count" db:"count"`
}

func (r *WarehouseModule) List() ([]*WarehouseModule, error) {
	rows := []*WarehouseModule{}
	err := r.Db.Select(&rows,
		`SELECT user_id, module_type, count
		FROM warehouse_modules
		WHERE user_id=?
		ORDER BY module_type ASC`,
		r.UserID,
	)
	return rows, err
}

func (r *WarehouseModule) Upsert(moduleType WarehouseModuleType, count int) (*WarehouseModule, error) {
	if count < 0 {
		count = 0
	}

	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		`INSERT INTO warehouse_modules (user_id, module_type, count)
		VALUES (:userID, :moduleType, :count)
		ON DUPLICATE KEY UPDATE count=:count`,
		map[string]interface{}{
			"userID":     r.UserID,
			"moduleType": strings.TrimSpace(moduleType.String()),
			"count":      count,
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
		`SELECT user_id, module_type, count
		FROM warehouse_modules
		WHERE user_id=? AND module_type=?
		LIMIT 1`,
		r.UserID, moduleType,
	)
	if err != nil {
		return nil, err
	}

	return &current, nil
}
