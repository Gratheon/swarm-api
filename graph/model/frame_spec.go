package model

import (
	"github.com/jmoiron/sqlx"
)

type FrameSpec struct {
	Db          *sqlx.DB
	UserID      string
	ID          string    `json:"id" db:"id"`
	SystemID    string    `json:"systemId" db:"system_id"`
	Code        string    `json:"code" db:"code"`
	FrameType   FrameType `json:"frameType" db:"frame_type"`
	DisplayName string    `json:"displayName" db:"display_name"`
	Active      bool      `json:"active" db:"active"`
}

func (r *FrameSpec) ListVisible(systemID *string) ([]*FrameSpec, error) {
	rows := []*FrameSpec{}
	query := ""
	args := []interface{}{}

	if systemID != nil && *systemID != "" {
		query = `
			SELECT DISTINCT
				CAST(fs.id AS CHAR) AS id,
				CAST(fs.system_id AS CHAR) AS system_id,
				fs.code,
				fs.frame_type,
				fs.display_name,
				fs.active
			FROM box_specs target_bs
			INNER JOIN box_systems target_sys ON target_sys.id = target_bs.system_id AND target_sys.active = 1
			INNER JOIN frame_spec_compatibility c ON c.box_spec_id = target_bs.id
			INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id AND fs.active = 1
			INNER JOIN box_systems source_sys ON source_sys.id = fs.system_id AND source_sys.active = 1
			WHERE target_bs.active = 1
			  AND target_bs.system_id = ?
			  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
			  AND (target_sys.user_id = ? OR target_sys.user_id IS NULL)
			  AND (source_sys.user_id = ? OR source_sys.user_id IS NULL)
			ORDER BY fs.display_name ASC, fs.id ASC
		`
		args = []interface{}{*systemID, r.UserID, r.UserID}
	} else {
		query = `
			SELECT
				CAST(fs.id AS CHAR) AS id,
				CAST(fs.system_id AS CHAR) AS system_id,
				fs.code,
				fs.frame_type,
				fs.display_name,
				fs.active
			FROM frame_specs fs
			INNER JOIN box_systems bs ON bs.id = fs.system_id AND bs.active = 1
			WHERE fs.active = 1
			  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
			  AND (bs.user_id = ? OR bs.user_id IS NULL)
			ORDER BY bs.name ASC, fs.display_name ASC, fs.id ASC
		`
		args = []interface{}{r.UserID}
	}
	err := r.Db.Select(&rows, query, args...)
	return rows, err
}
