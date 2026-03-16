package model

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var errNoCompatibleFrameSpec = errors.New("frame type is not compatible with target box specification")

type boxSpecLookupRow struct {
	BoxSystemID int `db:"box_system_id"`
	BoxSpecID   int `db:"box_spec_id"`
}

func getDefaultBoxSpecForType(db sqlx.QueryerContext, userID string, boxType BoxType) (*boxSpecLookupRow, error) {
	row := boxSpecLookupRow{}
	err := sqlx.GetContext(
		context.Background(),
		db,
		&row,
		`SELECT
			sys.id AS box_system_id,
			spec.id AS box_spec_id
		FROM box_systems sys
		INNER JOIN box_specs spec ON spec.system_id = sys.id
		WHERE spec.legacy_box_type = ?
		  AND spec.active = 1
		  AND sys.active = 1
		  AND (sys.user_id = ? OR sys.user_id IS NULL)
		ORDER BY (sys.user_id IS NULL) ASC, sys.is_default DESC, sys.id ASC
		LIMIT 1`,
		boxType.String(),
		userID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func getBoxSpecForTypeInSystem(db sqlx.QueryerContext, userID string, systemID int, boxType BoxType) (*boxSpecLookupRow, error) {
	row := boxSpecLookupRow{}
	err := sqlx.GetContext(
		context.Background(),
		db,
		&row,
		`SELECT
			sys.id AS box_system_id,
			spec.id AS box_spec_id
		FROM box_systems sys
		INNER JOIN box_specs spec ON spec.system_id = sys.id
		WHERE sys.id = ?
		  AND spec.legacy_box_type = ?
		  AND spec.active = 1
		  AND sys.active = 1
		  AND (sys.user_id = ? OR sys.user_id IS NULL)
		LIMIT 1`,
		systemID,
		boxType.String(),
		userID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func resolveFrameSpecForTargetBox(db sqlx.QueryerContext, userID string, boxID string, frameType FrameType) (int, error) {
	var frameSpecID int
	err := sqlx.GetContext(
		context.Background(),
		db,
		&frameSpecID,
		`SELECT fs.id
		FROM boxes b
		INNER JOIN frame_spec_compatibility c ON c.box_spec_id = b.box_spec_id
		INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
		WHERE b.id = ?
		  AND b.user_id = ?
		  AND b.active = 1
		  AND b.box_spec_id IS NOT NULL
		  AND fs.frame_type = ?
		  AND fs.active = 1
		ORDER BY fs.id ASC
		LIMIT 1`,
		boxID, userID, frameType.String(),
	)
	if err == sql.ErrNoRows {
		return 0, errNoCompatibleFrameSpec
	}
	return frameSpecID, err
}

func resolveEffectiveBoxProfileSystemID(db sqlx.QueryerContext, userID string, systemID int) (int, error) {
	current := systemID
	visited := map[int]bool{}

	for {
		if visited[current] {
			return 0, errors.New("box profile source cycle detected")
		}
		visited[current] = true

		var next sql.NullInt64
		err := sqlx.GetContext(
			context.Background(),
			db,
			&next,
			`SELECT box_profile_source_system_id
			FROM box_systems
			WHERE id = ?
			  AND active = 1
			  AND (user_id = ? OR user_id IS NULL)
			LIMIT 1`,
			current,
			userID,
		)
		if err != nil {
			return 0, err
		}
		if !next.Valid || int(next.Int64) <= 0 {
			return current, nil
		}
		current = int(next.Int64)
	}
}
