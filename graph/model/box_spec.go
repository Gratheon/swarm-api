package model

import "github.com/jmoiron/sqlx"

type BoxSpec struct {
	Db            *sqlx.DB
	UserID        string
	ID            string  `json:"id" db:"id"`
	SystemID      string  `json:"systemId" db:"system_id"`
	Code          string  `json:"code" db:"code"`
	LegacyBoxType BoxType `json:"legacyBoxType" db:"legacy_box_type"`
	DisplayName   string  `json:"displayName" db:"display_name"`
	Active        bool    `json:"active" db:"active"`
}

func (r *BoxSpec) ListVisible(systemID string) ([]*BoxSpec, error) {
	rows := []*BoxSpec{}
	err := r.Db.Select(&rows, `
		SELECT
			CAST(bs.id AS CHAR) AS id,
			CAST(bs.system_id AS CHAR) AS system_id,
			bs.code,
			bs.legacy_box_type,
			bs.display_name,
			bs.active
		FROM box_specs bs
		INNER JOIN box_systems sys ON sys.id = bs.system_id AND sys.active = 1
		WHERE bs.active = 1
		  AND bs.system_id = ?
		  AND (sys.user_id = ? OR sys.user_id IS NULL)
		ORDER BY bs.id ASC
	`, systemID, r.UserID)
	return rows, err
}
