package model

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type BoxSpec struct {
	Db               *sqlx.DB
	UserID           string
	ID               string  `json:"id" db:"id"`
	SystemID         string  `json:"systemId" db:"system_id"`
	Code             string  `json:"code" db:"code"`
	LegacyBoxType    BoxType `json:"legacyBoxType" db:"legacy_box_type"`
	DisplayName      string  `json:"displayName" db:"display_name"`
	InternalWidthMM  *int    `json:"internalWidthMm" db:"internal_width_mm"`
	InternalLengthMM *int    `json:"internalLengthMm" db:"internal_length_mm"`
	InternalHeightMM *int    `json:"internalHeightMm" db:"internal_height_mm"`
	ExternalWidthMM  *int    `json:"externalWidthMm" db:"external_width_mm"`
	ExternalLengthMM *int    `json:"externalLengthMm" db:"external_length_mm"`
	FrameWidthMM     *int    `json:"frameWidthMm" db:"frame_width_mm"`
	FrameHeightMM    *int    `json:"frameHeightMm" db:"frame_height_mm"`
	Active           bool    `json:"active" db:"active"`
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
			bs.internal_width_mm,
			bs.internal_length_mm,
			bs.internal_height_mm,
			bs.external_width_mm,
			bs.external_length_mm,
			bs.frame_width_mm,
			bs.frame_height_mm,
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

func (r *BoxSpec) SetDimensionsBySystemAndType(
	systemID string,
	boxType BoxType,
	internalWidthMM *int,
	internalLengthMM *int,
	internalHeightMM *int,
	externalWidthMM *int,
	externalLengthMM *int,
	frameWidthMM *int,
	frameHeightMM *int,
) (bool, error) {
	targetSystemID := strings.TrimSpace(systemID)
	if targetSystemID == "" {
		return false, fmt.Errorf("system id is required")
	}
	if boxType != BoxTypeDeep && boxType != BoxTypeSuper && boxType != BoxTypeLargeHorizontalSection {
		return false, fmt.Errorf("box type is not configurable")
	}

	args := []*int{
		internalWidthMM,
		internalLengthMM,
		internalHeightMM,
		externalWidthMM,
		externalLengthMM,
		frameWidthMM,
		frameHeightMM,
	}
	for _, value := range args {
		if value != nil && *value < 0 {
			return false, fmt.Errorf("dimensions must be non-negative")
		}
	}
	if internalWidthMM != nil && frameWidthMM != nil && *frameWidthMM > *internalWidthMM {
		return false, fmt.Errorf("frame width cannot be greater than section internal width")
	}
	if internalHeightMM != nil && frameHeightMM != nil && *frameHeightMM > *internalHeightMM {
		return false, fmt.Errorf("frame height cannot be greater than section internal height")
	}

	result, err := r.Db.NamedExec(`
		UPDATE box_specs bs
		INNER JOIN box_systems sys ON sys.id = bs.system_id
			SET
				bs.internal_width_mm = :internal_width_mm,
				bs.internal_length_mm = :internal_length_mm,
				bs.internal_height_mm = :internal_height_mm,
				bs.external_width_mm = :external_width_mm,
				bs.external_length_mm = :external_length_mm,
				bs.frame_width_mm = :frame_width_mm,
				bs.frame_height_mm = :frame_height_mm
		WHERE bs.system_id = :system_id
		  AND bs.legacy_box_type = :box_type
		  AND bs.active = 1
		  AND sys.active = 1
		  AND sys.user_id = :user_id
	`, map[string]interface{}{
		"system_id":          targetSystemID,
		"box_type":           boxType.String(),
		"user_id":            r.UserID,
		"internal_width_mm":  internalWidthMM,
		"internal_length_mm": internalLengthMM,
		"internal_height_mm": internalHeightMM,
		"external_width_mm":  externalWidthMM,
		"external_length_mm": externalLengthMM,
		"frame_width_mm":     frameWidthMM,
		"frame_height_mm":    frameHeightMM,
	})
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, fmt.Errorf("box spec not found")
	}
	return true, nil
}
