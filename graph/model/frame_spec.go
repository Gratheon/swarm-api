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
	query := `
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
		  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID')
		  AND (bs.user_id = ? OR bs.user_id IS NULL)
	`
	args := []interface{}{r.UserID}

	if systemID != nil && *systemID != "" {
		query += " AND fs.system_id = ?"
		args = append(args, *systemID)
	}

	query += " ORDER BY bs.name ASC, fs.display_name ASC, fs.id ASC"
	err := r.Db.Select(&rows, query, args...)
	return rows, err
}
