package model

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type BoxSystemFrameSetting struct {
	Db                  *sqlx.DB
	UserID              string
	SystemID            string  `json:"systemId" db:"system_id"`
	BoxSpecID           string  `json:"boxSpecId" db:"box_spec_id"`
	BoxType             BoxType `json:"boxType" db:"box_type"`
	BoxDisplayName      string  `json:"boxDisplayName" db:"box_display_name"`
	FrameSourceSystemID *string `json:"frameSourceSystemId" db:"frame_source_system_id"`
}

func (r *BoxSystemFrameSetting) ListAllVisible() ([]*BoxSystemFrameSetting, error) {
	rows := []*BoxSystemFrameSetting{}
	err := r.Db.Select(&rows, `
		SELECT
			CAST(target_sys.id AS CHAR) AS system_id,
			CAST(bs.id AS CHAR) AS box_spec_id,
			bs.legacy_box_type AS box_type,
			bs.display_name AS box_display_name,
			CAST(COALESCE(cfg.frame_source_system_id, src.frame_source_system_id) AS CHAR) AS frame_source_system_id
		FROM box_specs bs
		INNER JOIN box_systems target_sys ON target_sys.id = bs.system_id AND target_sys.active = 1
		LEFT JOIN box_spec_frame_sources cfg ON cfg.box_spec_id = bs.id
		LEFT JOIN (
			SELECT c.box_spec_id, MIN(fs.system_id) AS frame_source_system_id
			FROM frame_spec_compatibility c
			INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
			WHERE fs.active = 1
			  AND fs.frame_type = 'FOUNDATION'
			GROUP BY c.box_spec_id
		) src ON src.box_spec_id = bs.id
		WHERE bs.active = 1
		  AND bs.legacy_box_type IN ('DEEP', 'SUPER', 'LARGE_HORIZONTAL_SECTION')
		  AND (target_sys.user_id = ? OR target_sys.user_id IS NULL)
		ORDER BY target_sys.name ASC, bs.id ASC
	`, r.UserID)
	return rows, err
}

func (r *BoxSystemFrameSetting) SetFrameSource(systemID string, boxType BoxType, frameSourceSystemID string) (bool, error) {
	targetSystemID := strings.TrimSpace(systemID)
	sourceSystemID := strings.TrimSpace(frameSourceSystemID)
	if targetSystemID == "" || sourceSystemID == "" {
		return false, fmt.Errorf("system ids are required")
	}
	if boxType != BoxTypeDeep && boxType != BoxTypeSuper && boxType != BoxTypeLargeHorizontalSection {
		return false, fmt.Errorf("box type is not configurable for frame compatibility")
	}

	var targetOwned int
	err := r.Db.Get(&targetOwned, `
		SELECT COUNT(*)
		FROM box_systems
		WHERE id = ?
		  AND user_id = ?
		  AND active = 1
	`, targetSystemID, r.UserID)
	if err != nil {
		return false, err
	}
	if targetOwned == 0 {
		return false, fmt.Errorf("target box system not found")
	}

	var sourceVisible int
	err = r.Db.Get(&sourceVisible, `
		SELECT COUNT(*)
		FROM box_systems
		WHERE id = ?
		  AND active = 1
		  AND (user_id = ? OR user_id IS NULL)
	`, sourceSystemID, r.UserID)
	if err != nil {
		return false, err
	}
	if sourceVisible == 0 {
		return false, fmt.Errorf("frame source system not found")
	}

	tx := r.Db.MustBegin()

	var targetBoxSpecID int
	err = tx.Get(&targetBoxSpecID, `
		SELECT id
		FROM box_specs
		WHERE system_id = ?
		  AND legacy_box_type = ?
		  AND active = 1
		LIMIT 1
	`, targetSystemID, boxType.String())
	if err != nil {
		tx.Rollback()
		return false, err
	}

	var sourceBoxSpecID int
	err = tx.Get(&sourceBoxSpecID, `
		SELECT id
		FROM box_specs
		WHERE system_id = ?
		  AND legacy_box_type = ?
		  AND active = 1
		LIMIT 1
	`, sourceSystemID, boxType.String())
	if err != nil {
		tx.Rollback()
		return false, err
	}

	if sourceSystemID == targetSystemID {
		_, err = tx.Exec(`
			INSERT INTO frame_specs (system_id, code, frame_type, display_name, active)
			SELECT ?, fs.code, fs.frame_type, fs.display_name, fs.active
			FROM frame_spec_compatibility c
			INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
			WHERE c.box_spec_id = ?
			  AND fs.active = 1
			  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
			ON DUPLICATE KEY UPDATE
			  display_name = VALUES(display_name),
			  active = VALUES(active)
		`, sourceSystemID, sourceBoxSpecID)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	sourceFrameSpecIDs := []int{}
	if sourceSystemID == targetSystemID {
		likeSuffix := "%_DEEP"
		if boxType == BoxTypeSuper {
			likeSuffix = "%_SUPER"
		}
		if boxType == BoxTypeLargeHorizontalSection {
			likeSuffix = "%_HORIZONTAL"
		}
		err = tx.Select(&sourceFrameSpecIDs, `
			SELECT fs.id
			FROM frame_specs fs
			WHERE fs.system_id = ?
			  AND fs.active = 1
			  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
			  AND fs.code LIKE ?
			ORDER BY fs.id ASC
		`, sourceSystemID, likeSuffix)
	} else {
		err = tx.Select(&sourceFrameSpecIDs, `
			SELECT DISTINCT fs.id
			FROM frame_spec_compatibility c
			INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
			WHERE c.box_spec_id = ?
			  AND fs.active = 1
			  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
			ORDER BY fs.id ASC
		`, sourceBoxSpecID)
	}
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if len(sourceFrameSpecIDs) == 0 {
		tx.Rollback()
		return false, fmt.Errorf("source box system has no compatible frame specs for selected box type")
	}

	_, err = tx.Exec(`
		DELETE c
		FROM frame_spec_compatibility c
		INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
		WHERE c.box_spec_id = ?
		  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
	`, targetBoxSpecID)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	for _, frameSpecID := range sourceFrameSpecIDs {
		_, err = tx.Exec(`
			INSERT INTO frame_spec_compatibility (frame_spec_id, box_spec_id)
			VALUES (?, ?)
		`, frameSpecID, targetBoxSpecID)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	_, err = tx.Exec(`
		INSERT INTO box_spec_frame_sources (box_spec_id, frame_source_system_id)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE frame_source_system_id = VALUES(frame_source_system_id)
	`, targetBoxSpecID, sourceSystemID)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (r *BoxSystemFrameSetting) ResolveFrameSourceSystemByBoxSpec(boxSpecID int) (*string, error) {
	var sourceSystemID sql.NullString
	err := r.Db.Get(&sourceSystemID, `
		SELECT CAST(fs.system_id AS CHAR)
		FROM frame_spec_compatibility c
		INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id AND fs.active = 1 AND fs.frame_type = 'FOUNDATION'
		WHERE c.box_spec_id = ?
		ORDER BY fs.id ASC
		LIMIT 1
	`, boxSpecID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !sourceSystemID.Valid || strings.TrimSpace(sourceSystemID.String) == "" {
		return nil, nil
	}
	value := sourceSystemID.String
	return &value, nil
}
