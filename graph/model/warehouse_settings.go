package model

import "database/sql"
import "github.com/jmoiron/sqlx"

type WarehouseSettings struct {
	Db *sqlx.DB

	UserID              string `json:"user_id" db:"user_id"`
	AutoUpdateFromHives bool   `json:"autoUpdateFromHives" db:"auto_update_from_hives"`
}

func (r *WarehouseSettings) Get() (*WarehouseSettings, error) {
	settings := WarehouseSettings{}
	err := r.Db.Get(&settings,
		`SELECT user_id, auto_update_from_hives
		FROM warehouse_settings
		WHERE user_id=?
		LIMIT 1`,
		r.UserID,
	)
	if err == nil {
		return &settings, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Return defaults when row does not exist yet.
	return &WarehouseSettings{
		UserID:              r.UserID,
		AutoUpdateFromHives: true,
	}, nil
}

func (r *WarehouseSettings) UpsertAutoUpdate(enabled bool) (*WarehouseSettings, error) {
	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		`INSERT INTO warehouse_settings (user_id, auto_update_from_hives)
		VALUES (:userID, :enabled)
		ON DUPLICATE KEY UPDATE auto_update_from_hives=:enabled`,
		map[string]interface{}{
			"userID":  r.UserID,
			"enabled": enabled,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	current := WarehouseSettings{}
	err = r.Db.Get(&current,
		`SELECT user_id, auto_update_from_hives
		FROM warehouse_settings
		WHERE user_id=?
		LIMIT 1`,
		r.UserID,
	)
	if err != nil {
		return nil, err
	}

	return &current, nil
}
