package model

import (
	"database/sql"
	"strconv"
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
	var count int
	err := r.Db.Get(&count,
		`SELECT count
		FROM warehouse_modules
		WHERE user_id=? AND module_type=?
		LIMIT 1`,
		r.UserID, moduleType)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func (r *WarehouseModule) UsageStats(moduleType WarehouseModuleType) (*WarehouseModuleStats, error) {
	availableCount, err := r.GetCountByType(moduleType)
	if err != nil {
		return nil, err
	}

	inUseCount := 0
	topHives := []*WarehouseModuleHiveUsage{}

	kind, mappedType := mapModuleTypeToHiveSource(moduleType)
	switch kind {
	case "box":
		inUseCount, err = r.countBoxesInUse(mappedType)
		if err != nil {
			return nil, err
		}

		topHives, err = r.topHivesByBoxType(mappedType, 10)
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

func (r *WarehouseModule) countBoxesInUse(boxType string) (int, error) {
	var count int
	err := r.Db.Get(&count,
		`SELECT COUNT(*) AS cnt
		FROM boxes b
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		WHERE b.user_id = ?
		  AND b.active = 1
		  AND b.type = ?
		  AND h.active = 1
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`,
		r.UserID, boxType)
	return count, err
}

func (r *WarehouseModule) countFramesInUse(frameType string) (int, error) {
	var count int
	err := r.Db.Get(&count,
		`SELECT COUNT(*) AS cnt
		FROM frames f
		INNER JOIN boxes b ON b.id = f.box_id AND b.user_id = f.user_id AND b.active = 1
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		WHERE f.user_id = ?
		  AND f.active = 1
		  AND f.type = ?
		  AND h.active = 1
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL`,
		r.UserID, frameType)
	return count, err
}

func (r *WarehouseModule) topHivesByBoxType(boxType string, limit int) ([]*WarehouseModuleHiveUsage, error) {
	rows := []warehouseUsageScanRow{}
	err := r.Db.Select(&rows,
		`SELECT
			h.id AS hive_id,
			h.hive_number AS hive_number,
			h.apiary_id AS apiary_id,
			a.name AS apiary_name,
			COUNT(*) AS usage_count
		FROM boxes b
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		LEFT JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = h.user_id AND a.active = 1
		WHERE b.user_id = ?
		  AND b.active = 1
		  AND b.type = ?
		  AND h.active = 1
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL
		GROUP BY h.id, h.hive_number, h.apiary_id, a.name
		ORDER BY usage_count DESC, h.hive_number IS NULL ASC, h.hive_number ASC, h.id ASC
		LIMIT ?`,
		r.UserID, boxType, limit)
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
		INNER JOIN boxes b ON b.id = f.box_id AND b.user_id = f.user_id AND b.active = 1
		INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
		LEFT JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = h.user_id AND a.active = 1
		WHERE f.user_id = ?
		  AND f.active = 1
		  AND f.type = ?
		  AND h.active = 1
		  AND h.collapse_date IS NULL
		  AND h.merged_into_hive_id IS NULL
		GROUP BY h.id, h.hive_number, h.apiary_id, a.name
		ORDER BY usage_count DESC, h.hive_number IS NULL ASC, h.hive_number ASC, h.id ASC
		LIMIT ?`,
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
