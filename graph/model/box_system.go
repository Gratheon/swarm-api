package model

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type BoxSystem struct {
	Db        *sqlx.DB
	ID        string `json:"id" db:"id"`
	UserID    string `json:"user_id" db:"user_id"`
	Name      string `json:"name" db:"name"`
	IsDefault bool   `json:"isDefault" db:"is_default"`
	Active    bool   `json:"active" db:"active"`
}

func (r *BoxSystem) ListVisible() ([]*BoxSystem, error) {
	rows := []*BoxSystem{}
	err := r.Db.Select(&rows, `
		SELECT
			CAST(id AS CHAR) AS id,
			COALESCE(user_id, '') AS user_id,
			name,
			is_default,
			active
		FROM box_systems
		WHERE active = 1
		  AND (user_id = ? OR user_id IS NULL)
		ORDER BY (user_id IS NULL) ASC, is_default DESC, id ASC
	`, r.UserID)
	return rows, err
}

func (r *BoxSystem) ResolveForCreate(requestedID *string) (int, error) {
	if requestedID != nil && strings.TrimSpace(*requestedID) != "" {
		var id int
		err := r.Db.Get(&id, `
			SELECT id
			FROM box_systems
			WHERE id = ?
			  AND active = 1
			  AND (user_id = ? OR user_id IS NULL)
			LIMIT 1
		`, *requestedID, r.UserID)
		if err == nil {
			return id, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	var id int
	err := r.Db.Get(&id, `
		SELECT id
		FROM box_systems
		WHERE active = 1
		  AND (user_id = ? OR user_id IS NULL)
		ORDER BY (user_id IS NULL) ASC, is_default DESC, id ASC
		LIMIT 1
	`, r.UserID)
	return id, err
}

func (r *BoxSystem) Create(name string) (*BoxSystem, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, errors.New("name is required")
	}

	var ownedCount int
	err := r.Db.Get(&ownedCount, `
		SELECT COUNT(*)
		FROM box_systems
		WHERE user_id = ?
		  AND active = 1
	`, r.UserID)
	if err != nil {
		return nil, err
	}
	if ownedCount >= 4 {
		return nil, errors.New("maximum 4 box systems allowed")
	}

	tx := r.Db.MustBegin()
	result, err := tx.NamedExec(`
		INSERT INTO box_systems (user_id, name, is_default, active)
		VALUES (:user_id, :name, 0, 1)
	`, map[string]interface{}{
		"user_id": r.UserID,
		"name":    trimmed,
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	systemID64, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	systemID := int(systemID64)

	var sourceID int
	err = tx.Get(&sourceID, `
		SELECT id
		FROM box_systems
		WHERE user_id IS NULL
		  AND name = 'Langstroth'
		  AND active = 1
		ORDER BY is_default DESC, id ASC
		LIMIT 1
	`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO box_specs (system_id, code, legacy_box_type, display_name, active)
		SELECT ?, code, legacy_box_type, display_name, active
		FROM box_specs
		WHERE system_id = ?
	`, systemID, sourceID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO frame_specs (system_id, code, frame_type, display_name, active)
		SELECT ?, code, frame_type, display_name, active
		FROM frame_specs
		WHERE system_id = ?
		  AND frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID')
	`, systemID, sourceID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO frame_spec_compatibility (frame_spec_id, box_spec_id)
		SELECT f2.id, b2.id
		FROM frame_spec_compatibility c
		INNER JOIN frame_specs f1 ON f1.id = c.frame_spec_id AND f1.system_id = ?
		INNER JOIN box_specs b1 ON b1.id = c.box_spec_id AND b1.system_id = ?
		INNER JOIN frame_specs f2 ON f2.system_id = ? AND f2.code = f1.code
		INNER JOIN box_specs b2 ON b2.system_id = ? AND b2.code = b1.code
	`, sourceID, sourceID, systemID, systemID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetOwnedByID(systemID)
}

func (r *BoxSystem) GetOwnedByID(id int) (*BoxSystem, error) {
	row := BoxSystem{}
	err := r.Db.Get(&row, `
		SELECT
			CAST(id AS CHAR) AS id,
			COALESCE(user_id, '') AS user_id,
			name,
			is_default,
			active
		FROM box_systems
		WHERE id = ?
		  AND user_id = ?
		  AND active = 1
		LIMIT 1
	`, id, r.UserID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BoxSystem) Rename(id string, name string) (*BoxSystem, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, errors.New("name is required")
	}

	var isDefault bool
	err := r.Db.Get(&isDefault, `
		SELECT is_default
		FROM box_systems
		WHERE id = ?
		  AND user_id = ?
		  AND active = 1
		LIMIT 1
	`, id, r.UserID)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if isDefault {
		return nil, errors.New("default box system cannot be renamed")
	}
	_, err = r.Db.NamedExec(`
		UPDATE box_systems
		SET name = :name
		WHERE id = :id
		  AND user_id = :user_id
		  AND active = 1
	`, map[string]interface{}{
		"id":      id,
		"user_id": r.UserID,
		"name":    trimmed,
	})
	if err != nil {
		return nil, err
	}
	idNum, convErr := strconv.Atoi(id)
	if convErr != nil {
		return nil, convErr
	}
	return r.GetOwnedByID(idNum)
}

func (r *BoxSystem) Deactivate(id string) (bool, error) {
	var isDefault bool
	err := r.Db.Get(&isDefault, `
		SELECT is_default
		FROM box_systems
		WHERE id = ?
		  AND user_id = ?
		  AND active = 1
		LIMIT 1
	`, id, r.UserID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if isDefault {
		return false, errors.New("default box system cannot be deactivated")
	}

	tx := r.Db.MustBegin()
	_, err = tx.NamedExec(`
		UPDATE box_systems
		SET active = 0
		WHERE id = :id
		  AND user_id = :user_id
		  AND active = 1
	`, map[string]interface{}{
		"id":      id,
		"user_id": r.UserID,
	})
	if err != nil {
		tx.Rollback()
		return false, err
	}

	_, err = tx.NamedExec(`
		UPDATE box_specs
		SET active = 0
		WHERE system_id = :id
	`, map[string]interface{}{"id": id})
	if err != nil {
		tx.Rollback()
		return false, err
	}

	_, err = tx.NamedExec(`
		UPDATE frame_specs
		SET active = 0
		WHERE system_id = :id
	`, map[string]interface{}{"id": id})
	if err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
