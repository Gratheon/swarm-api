package model

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type BoxSystem struct {
	Db                       *sqlx.DB
	ID                       string  `json:"id" db:"id"`
	UserID                   string  `json:"user_id" db:"user_id"`
	Name                     string  `json:"name" db:"name"`
	BoxProfileSourceSystemID *string `json:"boxProfileSourceSystemId" db:"box_profile_source_system_id"`
	IsDefault                bool    `json:"isDefault" db:"is_default"`
	Active                   bool    `json:"active" db:"active"`
}

func (r *BoxSystem) ListVisible() ([]*BoxSystem, error) {
	rows := []*BoxSystem{}
	err := r.Db.Select(&rows, `
		SELECT
			CAST(id AS CHAR) AS id,
			COALESCE(user_id, '') AS user_id,
			name,
			CAST(box_profile_source_system_id AS CHAR) AS box_profile_source_system_id,
			is_default,
			active
		FROM box_systems
		WHERE active = 1
		  AND (
			user_id = ?
			OR (
				user_id IS NULL
				AND (
					is_default = 0
					OR NOT EXISTS (
						SELECT 1
						FROM box_systems owned_defaults
						WHERE owned_defaults.user_id = ?
						  AND owned_defaults.active = 1
						  AND owned_defaults.is_default = 1
					)
				)
			)
		  )
		ORDER BY (user_id IS NULL) ASC, is_default DESC, id ASC
	`, r.UserID, r.UserID)
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
	if ownedCount >= 5 {
		return nil, errors.New("maximum 5 box systems allowed")
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
		INSERT INTO box_specs (
			system_id,
			code,
			legacy_box_type,
			display_name,
			internal_width_mm,
			internal_length_mm,
			internal_height_mm,
			external_width_mm,
			external_length_mm,
			frame_width_mm,
			frame_height_mm,
			active
		)
		SELECT
			?,
			code,
			legacy_box_type,
				display_name,
				internal_width_mm,
				internal_length_mm,
				internal_height_mm,
				external_width_mm,
				external_length_mm,
				frame_width_mm,
				frame_height_mm,
				active
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
		  AND active = 1
	`, systemID, sourceID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO frame_spec_compatibility (frame_spec_id, box_spec_id)
		SELECT f2.id, b2.id
		FROM frame_spec_compatibility c
		INNER JOIN box_specs b1 ON b1.id = c.box_spec_id AND b1.system_id = ?
		INNER JOIN frame_specs f1 ON f1.id = c.frame_spec_id AND f1.system_id = ? AND f1.active = 1
		INNER JOIN frame_specs f2 ON f2.system_id = ? AND f2.code = f1.code AND f2.active = 1
		INNER JOIN box_specs b2 ON b2.system_id = ? AND b2.code = b1.code
	`, sourceID, sourceID, systemID, systemID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO box_spec_frame_sources (box_spec_id, frame_source_system_id)
		SELECT b2.id, ?
		FROM box_specs b2
		WHERE b2.system_id = ?
		  AND b2.active = 1
		  AND b2.legacy_box_type IN ('DEEP', 'SUPER', 'LARGE_HORIZONTAL_SECTION')
		ON DUPLICATE KEY UPDATE frame_source_system_id = VALUES(frame_source_system_id)
	`, sourceID, systemID)
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
			CAST(box_profile_source_system_id AS CHAR) AS box_profile_source_system_id,
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

	var row struct {
		UserID    sql.NullString `db:"user_id"`
		IsDefault bool           `db:"is_default"`
	}
	err := r.Db.Get(&row, `
		SELECT user_id, is_default
		FROM box_systems
		WHERE id = ?
		  AND active = 1
		  AND (user_id = ? OR user_id IS NULL)
		LIMIT 1
	`, id, r.UserID)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if !row.UserID.Valid {
		if !row.IsDefault {
			return nil, errors.New("box system is not editable")
		}

		var existingOwnedDefaultID int
		existingOwnedDefaultErr := r.Db.Get(&existingOwnedDefaultID, `
			SELECT id
			FROM box_systems
			WHERE user_id = ?
			  AND is_default = 1
			  AND active = 1
			ORDER BY id ASC
			LIMIT 1
		`, r.UserID)
		if existingOwnedDefaultErr == nil {
			_, err = r.Db.NamedExec(`
				UPDATE box_systems
				SET name = :name
				WHERE id = :id
				  AND user_id = :user_id
				  AND active = 1
			`, map[string]interface{}{
				"id":      existingOwnedDefaultID,
				"user_id": r.UserID,
				"name":    trimmed,
			})
			if err != nil {
				return nil, err
			}
			return r.GetOwnedByID(existingOwnedDefaultID)
		}
		if existingOwnedDefaultErr != nil && existingOwnedDefaultErr != sql.ErrNoRows {
			return nil, existingOwnedDefaultErr
		}

		created, createErr := r.Create(trimmed)
		if createErr != nil {
			return nil, createErr
		}

		_, err = r.Db.NamedExec(`
			UPDATE box_systems
			SET is_default = 1
			WHERE id = :id
			  AND user_id = :user_id
			  AND active = 1
		`, map[string]interface{}{
			"id":      created.ID,
			"user_id": r.UserID,
		})
		if err != nil {
			return nil, err
		}

		idNum, convErr := strconv.Atoi(created.ID)
		if convErr != nil {
			return nil, convErr
		}
		return r.GetOwnedByID(idNum)
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

func (r *BoxSystem) Deactivate(id string, replacementSystemID *string) (bool, error) {
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

	replacementID := strings.TrimSpace(func() string {
		if replacementSystemID == nil {
			return ""
		}
		return *replacementSystemID
	}())

	if replacementID != "" {
		var replacementExists int
		err = r.Db.Get(&replacementExists, `
			SELECT COUNT(*)
			FROM box_systems
			WHERE id = ?
			  AND active = 1
			  AND (user_id = ? OR user_id IS NULL)
		`, replacementID, r.UserID)
		if err != nil {
			return false, err
		}
		if replacementExists == 0 {
			return false, errors.New("replacement box system not found")
		}
		if replacementID == id {
			return false, errors.New("replacement box system must be different")
		}
	}

	tx := r.Db.MustBegin()
	var hivesUsingSystem int
	err = tx.Get(&hivesUsingSystem, `
		SELECT COUNT(*)
		FROM hives
		WHERE user_id = ?
		  AND active = 1
		  AND box_system_id = ?
	`, r.UserID, id)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if hivesUsingSystem > 0 && replacementID == "" {
		tx.Rollback()
		return false, errors.New("replacement box system is required for active hives using this system")
	}
	if hivesUsingSystem > 0 {
		_, err = tx.Exec(`
			UPDATE hives
			SET box_system_id = ?
			WHERE user_id = ?
			  AND active = 1
			  AND box_system_id = ?
		`, replacementID, r.UserID, id)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

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

func (r *BoxSystem) SetBoxProfileSource(systemID string, boxSourceSystemID *string) (bool, error) {
	targetID := strings.TrimSpace(systemID)
	if targetID == "" {
		return false, errors.New("system id is required")
	}

	var isDefault bool
	err := r.Db.Get(&isDefault, `
		SELECT is_default
		FROM box_systems
		WHERE id = ?
		  AND user_id = ?
		  AND active = 1
		LIMIT 1
	`, targetID, r.UserID)
	if err == sql.ErrNoRows {
		return false, errors.New("box system not found")
	}
	if err != nil {
		return false, err
	}
	if isDefault {
		return false, errors.New("default box system profile cannot be changed")
	}

	sourceID := strings.TrimSpace(func() string {
		if boxSourceSystemID == nil {
			return ""
		}
		return *boxSourceSystemID
	}())

	// Own profile
	if sourceID == "" || sourceID == targetID {
		_, err = r.Db.NamedExec(`
			UPDATE box_systems
			SET box_profile_source_system_id = NULL
			WHERE id = :id
			  AND user_id = :user_id
			  AND active = 1
		`, map[string]interface{}{
			"id":      targetID,
			"user_id": r.UserID,
		})
		if err != nil {
			return false, err
		}
		return true, nil
	}

	var sourceVisible int
	err = r.Db.Get(&sourceVisible, `
		SELECT COUNT(*)
		FROM box_systems
		WHERE id = ?
		  AND active = 1
		  AND (user_id = ? OR user_id IS NULL)
	`, sourceID, r.UserID)
	if err != nil {
		return false, err
	}
	if sourceVisible == 0 {
		return false, errors.New("box source system not found")
	}

	visited := map[string]bool{targetID: true}
	current := sourceID
	for current != "" {
		if visited[current] {
			return false, errors.New("box profile source relationship would create a cycle")
		}
		visited[current] = true

		var next sql.NullString
		err = r.Db.Get(&next, `
			SELECT CAST(box_profile_source_system_id AS CHAR)
			FROM box_systems
			WHERE id = ?
			  AND active = 1
			  AND (user_id = ? OR user_id IS NULL)
			LIMIT 1
		`, current, r.UserID)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			return false, err
		}
		if !next.Valid || strings.TrimSpace(next.String) == "" {
			break
		}
		current = strings.TrimSpace(next.String)
	}

	_, err = r.Db.NamedExec(`
		UPDATE box_systems
		SET box_profile_source_system_id = :source_id
		WHERE id = :id
		  AND user_id = :user_id
		  AND active = 1
	`, map[string]interface{}{
		"id":        targetID,
		"user_id":   r.UserID,
		"source_id": sourceID,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
