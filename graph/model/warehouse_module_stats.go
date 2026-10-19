package model

import (
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type WarehouseModuleHiveUsage struct {
	HiveID     string  `json:"hiveId"`
	HiveNumber *int    `json:"hiveNumber"`
	ApiaryID   *string `json:"apiaryId"`
	ApiaryName *string `json:"apiaryName"`
	Count      int     `json:"count"`
}

type WarehouseModuleStats struct {
	ModuleType     WarehouseModuleType         `json:"moduleType"`
	AvailableCount int                         `json:"availableCount"`
	InUseCount     int                         `json:"inUseCount"`
	TotalCount     int                         `json:"totalCount"`
	TopHives       []*WarehouseModuleHiveUsage `json:"topHives"`
}

type warehouseUsageScanRow struct {
	HiveID     int            `db:"hive_id"`
	HiveNumber sql.NullInt64  `db:"hive_number"`
	ApiaryID   sql.NullInt64  `db:"apiary_id"`
	ApiaryName sql.NullString `db:"apiary_name"`
	Count      int            `db:"usage_count"`
}

func (r *WarehouseModule) GetCountByType(moduleType WarehouseModuleType) (int, error) {
	return r.GetCountByTypeAndSystem(moduleType, nil)
}

func (r *WarehouseModule) GetCountByTypeAndSystem(moduleType WarehouseModuleType, boxSystemID *int) (int, error) {
	systemID := 0
	if boxSystemID != nil && *boxSystemID > 0 {
		systemID = *boxSystemID
	}

	var count int
	err := r.Db.Get(&count,
		`SELECT count
		FROM warehouse_modules
		WHERE user_id=$1 AND module_type=$2 AND box_system_id=$3
		LIMIT 1`,
		r.UserID, moduleType, systemID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func (r *WarehouseModule) UsageStats(moduleType WarehouseModuleType) (*WarehouseModuleStats, error) {
	return r.UsageStatsForSystem(moduleType, nil)
}

func (r *WarehouseModule) UsageStatsForSystem(moduleType WarehouseModuleType, boxSystemID *int) (*WarehouseModuleStats, error) {
	availableCount, err := r.GetCountByTypeAndSystem(moduleType, boxSystemID)
	if err != nil {
		return nil, err
	}

	inUseCount := 0
	topHives := []*WarehouseModuleHiveUsage{}

	kind, mappedType := mapModuleTypeToHiveSource(moduleType)
	switch kind {
	case "box":
		inUseCount, err = r.countBoxesInUse(mappedType, boxSystemID)
		if err != nil {
			return nil, err
		}

		topHives, err = r.topHivesByBoxType(mappedType, boxSystemID, 10)
		if err != nil {
			return nil, err
		}
	case "frame":
		inUseCount, err = r.countFramesInUse(mappedType)
		if err != nil {
			return nil, err
		}

		topHives, err = r.topHivesByFrameType(mappedType, 10)
		if err != nil {
			return nil, err
		}
	case "nucleus":
		inUseCount, err = r.countNucleusHivesInUse(boxSystemID)
		if err != nil {
			return nil, err
		}
		topHives, err = r.topNucleusHives(boxSystemID, 10)
		if err != nil {
			return nil, err
		}
	}

	return &WarehouseModuleStats{
		ModuleType:     moduleType,
		AvailableCount: availableCount,
		InUseCount:     inUseCount,
		TotalCount:     availableCount + inUseCount,
		TopHives:       topHives,
	}, nil
}

func mapModuleTypeToHiveSource(moduleType WarehouseModuleType) (string, string) {
	switch moduleType {
	case WarehouseModuleTypeDeep:
		return "box", BoxTypeDeep.String()
	case WarehouseModuleTypeNucs:
		return "nucleus", "NUCLEUS"
	case WarehouseModuleTypeSuper:
		return "box", BoxTypeSuper.String()
	case WarehouseModuleTypeLargeHorizontalSection:
		return "box", BoxTypeLargeHorizontalSection.String()
	case WarehouseModuleTypeRoof:
		return "box", BoxTypeRoof.String()
	case WarehouseModuleTypeHorizontalFeeder:
		return "box", BoxTypeHorizontalFeeder.String()
	case WarehouseModuleTypeQueenExcluder:
		return "box", BoxTypeQueenExcluder.String()
	case WarehouseModuleTypeBottom:
		return "box", BoxTypeBottom.String()
	case WarehouseModuleTypeFrameFoundation:
		return "frame", FrameTypeFoundation.String()
	case WarehouseModuleTypeFrameEmptyComb:
		return "frame", FrameTypeEmptyComb.String()
	case WarehouseModuleTypeFramePartition:
		return "frame", FrameTypePartition.String()
	case WarehouseModuleTypeFrameFeeder:
		return "frame", FrameTypeFeeder.String()
	default:
		return "", ""
	}
}

func (r *WarehouseModule) countNucleusHivesInUse(boxSystemID *int) (int, error) {
	var count int
	query := `SELECT COUNT(*) AS cnt
		FROM hives h
		LEFT JOIN box_specs bs ON bs.system_id = h.box_system_id AND bs.legacy_box_type = 'DEEP' AND bs.active = TRUE
		LEFT JOIN box_spec_frame_sources bfs ON bfs.box_spec_id = bs.id
		WHERE h.user_id = ?
		  AND h.active = TRUE
		  AND h.hive_type = 'NUCLEUS'
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`
	args := []interface{}{r.UserID}
	if boxSystemID != nil && *boxSystemID > 0 {
		query += " AND COALESCE(bfs.frame_source_system_id, h.box_system_id) = ?"
		args = append(args, *boxSystemID)
	}
	err := r.Db.Get(&count, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	return count, err
}

func (r *WarehouseModule) topNucleusHives(boxSystemID *int, limit int) ([]*WarehouseModuleHiveUsage, error) {
	rows := []warehouseUsageScanRow{}
	query := `SELECT
			h.id AS hive_id,
			h.hive_number AS hive_number,
			h.apiary_id AS apiary_id,
			a.name AS apiary_name,
			1 AS usage_count
		FROM hives h
		LEFT JOIN box_specs bs ON bs.system_id = h.box_system_id AND bs.legacy_box_type = 'DEEP' AND bs.active = TRUE
		LEFT JOIN box_spec_frame_sources bfs ON bfs.box_spec_id = bs.id
		LEFT JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = h.user_id AND a.active = TRUE
		WHERE h.user_id = ?
		  AND h.active = TRUE
		  AND h.hive_type = 'NUCLEUS'
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`
	args := []interface{}{r.UserID}
	if boxSystemID != nil && *boxSystemID > 0 {
		query += " AND COALESCE(bfs.frame_source_system_id, h.box_system_id) = ?"
		args = append(args, *boxSystemID)
	}
	query += `
		ORDER BY h.hive_number IS NULL ASC, h.hive_number ASC, h.id ASC
		LIMIT ?`
	args = append(args, limit)
	err := r.Db.Select(&rows, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	if err != nil {
		return nil, err
	}
	return mapUsageRows(rows), nil
}

func (r *WarehouseModule) countBoxesInUse(boxType string, boxSystemID *int) (int, error) {
	var count int
	query := `SELECT COUNT(*) AS cnt
		FROM boxes b
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		WHERE b.user_id = ?
		  AND b.active = TRUE
		  AND b.type = ?
		  AND h.active = TRUE
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`
	args := []interface{}{r.UserID, boxType}
	if boxSystemID != nil && *boxSystemID > 0 {
		query += " AND h.box_system_id = ?"
		args = append(args, *boxSystemID)
	}
	err := r.Db.Get(&count, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	return count, err
}

func (r *WarehouseModule) countFramesInUse(frameType string) (int, error) {
	var count int
	err := r.Db.Get(&count,
		`SELECT COUNT(*) AS cnt
		FROM frames f
		INNER JOIN boxes b ON b.id = f.box_id AND b.user_id = f.user_id AND b.active = TRUE
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		WHERE f.user_id = $1
		  AND f.active = TRUE
		  AND f.type = $2
		  AND h.active = TRUE
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`,
		r.UserID, frameType)
	return count, err
}

func (r *WarehouseModule) topHivesByBoxType(boxType string, boxSystemID *int, limit int) ([]*WarehouseModuleHiveUsage, error) {
	rows := []warehouseUsageScanRow{}
	query := `SELECT
			h.id AS hive_id,
			h.hive_number AS hive_number,
			h.apiary_id AS apiary_id,
			a.name AS apiary_name,
			COUNT(*) AS usage_count
		FROM boxes b
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		LEFT JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = h.user_id AND a.active = TRUE
		WHERE b.user_id = ?
		  AND b.active = TRUE
		  AND b.type = ?
		  AND h.active = TRUE
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`
	args := []interface{}{r.UserID, boxType}
	if boxSystemID != nil && *boxSystemID > 0 {
		query += " AND h.box_system_id = ?"
		args = append(args, *boxSystemID)
	}
	query += `
		GROUP BY h.id, h.hive_number, h.apiary_id, a.name
		ORDER BY usage_count DESC, h.hive_number IS NULL ASC, h.hive_number ASC, h.id ASC
		LIMIT ?`
	args = append(args, limit)
	err := r.Db.Select(&rows, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	if err != nil {
		return nil, err
	}

	return mapUsageRows(rows), nil
}

func (r *WarehouseModule) topHivesByFrameType(frameType string, limit int) ([]*WarehouseModuleHiveUsage, error) {
	rows := []warehouseUsageScanRow{}
	err := r.Db.Select(&rows,
		`SELECT
			h.id AS hive_id,
			h.hive_number AS hive_number,
			h.apiary_id AS apiary_id,
			a.name AS apiary_name,
			COUNT(*) AS usage_count
		FROM frames f
		INNER JOIN boxes b ON b.id = f.box_id AND b.user_id = f.user_id AND b.active = TRUE
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		LEFT JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = h.user_id AND a.active = TRUE
		WHERE f.user_id = $1
		  AND f.active = TRUE
		  AND f.type = $2
		  AND h.active = TRUE
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL
		GROUP BY h.id, h.hive_number, h.apiary_id, a.name
		ORDER BY usage_count DESC, h.hive_number IS NULL ASC, h.hive_number ASC, h.id ASC
		LIMIT $3`,
		r.UserID, frameType, limit)
	if err != nil {
		return nil, err
	}

	return mapUsageRows(rows), nil
}

func mapUsageRows(rows []warehouseUsageScanRow) []*WarehouseModuleHiveUsage {
	out := make([]*WarehouseModuleHiveUsage, 0, len(rows))
	for _, row := range rows {
		hiveID := strconv.Itoa(row.HiveID)

		var hiveNumber *int
		if row.HiveNumber.Valid {
			value := int(row.HiveNumber.Int64)
			hiveNumber = &value
		}

		var apiaryID *string
		if row.ApiaryID.Valid {
			value := strconv.Itoa(int(row.ApiaryID.Int64))
			apiaryID = &value
		}

		var apiaryName *string
		if row.ApiaryName.Valid {
			value := row.ApiaryName.String
			apiaryName = &value
		}

		out = append(out, &WarehouseModuleHiveUsage{
			HiveID:     hiveID,
			HiveNumber: hiveNumber,
			ApiaryID:   apiaryID,
			ApiaryName: apiaryName,
			Count:      row.Count,
		})
	}
	return out
}
