package model

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	warehouseItemKeyPrefixBox       = "BOX:"
	warehouseItemKeyPrefixFrameSpec = "FRAME_SPEC:"
	warehouseItemKeySystemDelimiter = ":SYSTEM:"
)

type WarehouseInventoryItemKind string

const (
	WarehouseInventoryItemKindBoxModule WarehouseInventoryItemKind = "BOX_MODULE"
	WarehouseInventoryItemKindFrameSpec WarehouseInventoryItemKind = "FRAME_SPEC"
)

type WarehouseInventoryItem struct {
	Key         string                     `json:"key"`
	Kind        WarehouseInventoryItemKind `json:"kind"`
	Count       int                        `json:"count"`
	GroupKey    string                     `json:"groupKey"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	ModuleType  *WarehouseModuleType       `json:"moduleType"`
	FrameSpec   *FrameSpec                 `json:"frameSpec"`
}

type WarehouseInventoryStats struct {
	Key            string                      `json:"key"`
	AvailableCount int                         `json:"availableCount"`
	InUseCount     int                         `json:"inUseCount"`
	TotalCount     int                         `json:"totalCount"`
	TopHives       []*WarehouseModuleHiveUsage `json:"topHives"`
}

type WarehouseInventory struct {
	Db     *sqlx.DB
	UserID string
}

type warehouseFrameSpecRow struct {
	ID          int       `db:"id"`
	SystemID    int       `db:"system_id"`
	Code        string    `db:"code"`
	FrameType   FrameType `db:"frame_type"`
	DisplayName string    `db:"display_name"`
	Count       int       `db:"count"`
}

type warehouseSystemRow struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func (r *WarehouseInventory) List() ([]*WarehouseInventoryItem, error) {
	items := []*WarehouseInventoryItem{}

	moduleModel := &WarehouseModule{
		Db:     r.Db,
		UserID: r.UserID,
	}

	systems := []warehouseSystemRow{}
	err := r.Db.Select(&systems, `
		SELECT id, name
		FROM box_systems
		WHERE active = 1
		  AND (user_id = ? OR user_id IS NULL)
		ORDER BY (user_id IS NULL) ASC, is_default DESC, id ASC
	`, r.UserID)
	if err != nil {
		return nil, err
	}

	for _, moduleType := range AllWarehouseModuleType {
		if isLegacyFrameModuleType(moduleType) {
			continue
		}
		if isSystemScopedHivePartModuleType(moduleType) {
			for _, system := range systems {
				systemID := system.ID
				count, err := moduleModel.GetCountByTypeAndSystem(moduleType, &systemID)
				if err != nil {
					return nil, err
				}
				items = append(items, &WarehouseInventoryItem{
					Key:         buildBoxInventoryKey(moduleType, &systemID),
					Kind:        WarehouseInventoryItemKindBoxModule,
					Count:       count,
					GroupKey:    mapBoxModuleGroup(moduleType),
					Title:       mapBoxModuleTitle(moduleType),
					Description: mapBoxModuleDescription(moduleType),
					ModuleType:  &moduleType,
				})
			}
			continue
		}

		count, err := moduleModel.GetCountByType(moduleType)
		if err != nil {
			return nil, err
		}
		items = append(items, &WarehouseInventoryItem{
			Key:         buildBoxInventoryKey(moduleType, nil),
			Kind:        WarehouseInventoryItemKindBoxModule,
			Count:       count,
			GroupKey:    mapBoxModuleGroup(moduleType),
			Title:       mapBoxModuleTitle(moduleType),
			Description: mapBoxModuleDescription(moduleType),
			ModuleType:  &moduleType,
		})
	}

	frameRows := []warehouseFrameSpecRow{}
	err = r.Db.Select(&frameRows, `
		SELECT
			fs.id,
			fs.system_id,
			fs.code,
			fs.frame_type,
			fs.display_name,
			COALESCE(wfi.count, 0) AS count
		FROM frame_specs fs
		INNER JOIN box_systems bs ON bs.id = fs.system_id AND bs.active = 1
		LEFT JOIN warehouse_frame_inventory wfi ON wfi.frame_spec_id = fs.id AND wfi.user_id = ?
		WHERE fs.active = 1
		  AND fs.frame_type IN ('FOUNDATION', 'EMPTY_COMB', 'VOID', 'PARTITION', 'FEEDER')
		  AND (bs.user_id = ? OR bs.user_id IS NULL)
		ORDER BY bs.name ASC, fs.display_name ASC, fs.id ASC
	`, r.UserID, r.UserID)
	if err != nil {
		return nil, err
	}

	for _, row := range frameRows {
		specID := strconv.Itoa(row.ID)
		systemID := strconv.Itoa(row.SystemID)
		spec := &FrameSpec{
			ID:          specID,
			SystemID:    systemID,
			Code:        row.Code,
			FrameType:   row.FrameType,
			DisplayName: row.DisplayName,
			Active:      true,
		}
		items = append(items, &WarehouseInventoryItem{
			Key:         warehouseItemKeyPrefixFrameSpec + specID,
			Kind:        WarehouseInventoryItemKindFrameSpec,
			Count:       row.Count,
			GroupKey:    "FRAMES",
			Title:       row.DisplayName,
			Description: "Frames compatible with a specific hive section size and system.",
			FrameSpec:   spec,
		})
	}

	return items, nil
}

func (r *WarehouseInventory) UpsertByKey(itemKey string, count int) (*WarehouseInventoryItem, error) {
	if count < 0 {
		count = 0
	}

	if strings.HasPrefix(itemKey, warehouseItemKeyPrefixBox) {
		moduleType, systemID, err := parseBoxInventoryKey(itemKey)
		if err != nil {
			return nil, err
		}

		updated, err := (&WarehouseModule{
			Db:     r.Db,
			UserID: r.UserID,
		}).UpsertForSystem(moduleType, systemID, count)
		if err != nil {
			return nil, err
		}

		mt := updated.ModuleType
		return &WarehouseInventoryItem{
			Key:         buildBoxInventoryKey(mt, systemID),
			Kind:        WarehouseInventoryItemKindBoxModule,
			Count:       updated.Count,
			GroupKey:    mapBoxModuleGroup(mt),
			Title:       mapBoxModuleTitle(mt),
			Description: mapBoxModuleDescription(mt),
			ModuleType:  &mt,
		}, nil
	}

	if strings.HasPrefix(itemKey, warehouseItemKeyPrefixFrameSpec) {
		specIDStr := strings.TrimPrefix(itemKey, warehouseItemKeyPrefixFrameSpec)
		specID, err := strconv.Atoi(specIDStr)
		if err != nil || specID <= 0 {
			return nil, fmt.Errorf("invalid frame spec key: %s", itemKey)
		}

		tx := r.Db.MustBegin()
		_, err = tx.NamedExec(`
			INSERT INTO warehouse_frame_inventory (user_id, frame_spec_id, count)
			VALUES (:user_id, :frame_spec_id, :count)
			ON DUPLICATE KEY UPDATE count=:count
		`, map[string]interface{}{
			"user_id":       r.UserID,
			"frame_spec_id": specID,
			"count":         count,
		})
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}

		items, err := r.List()
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.Key == itemKey {
				return item, nil
			}
		}
		return nil, fmt.Errorf("updated item not found: %s", itemKey)
	}

	return nil, fmt.Errorf("unsupported warehouse inventory key: %s", itemKey)
}

func (r *WarehouseInventory) UpdateFrameSpecByDelta(frameSpecID int, delta int) (*WarehouseInventoryItem, error) {
	var current int
	err := r.Db.Get(&current, `
		SELECT count
		FROM warehouse_frame_inventory
		WHERE user_id=? AND frame_spec_id=?
		LIMIT 1
	`, r.UserID, frameSpecID)
	if err == sql.ErrNoRows {
		current = 0
	} else if err != nil {
		return nil, err
	}

	next := current + delta
	if next < 0 {
		next = 0
	}
	return r.UpsertByKey(warehouseItemKeyPrefixFrameSpec+strconv.Itoa(frameSpecID), next)
}

func (r *WarehouseInventory) ResolveFrameSpecIDByBoxAndFrameType(boxID string, frameType FrameType) (int, error) {
	return resolveFrameSpecForTargetBox(r.Db, r.UserID, boxID, frameType)
}

func (r *WarehouseInventory) ResolveFrameSpecIDByFrameID(frameID string) (int, error) {
	var frameSpecID int
	err := r.Db.Get(&frameSpecID, `
		SELECT frame_spec_id
		FROM frames
		WHERE id=? AND user_id=? AND active=1
		LIMIT 1
	`, frameID, r.UserID)
	if err != nil {
		return 0, err
	}
	return frameSpecID, nil
}

func (r *WarehouseInventory) StatsByKey(itemKey string) (*WarehouseInventoryStats, error) {
	if strings.HasPrefix(itemKey, warehouseItemKeyPrefixBox) {
		moduleType, systemID, err := parseBoxInventoryKey(itemKey)
		if err != nil {
			return nil, err
		}
		stats, err := (&WarehouseModule{
			Db:     r.Db,
			UserID: r.UserID,
		}).UsageStatsForSystem(moduleType, systemID)
		if err != nil {
			return nil, err
		}
		return &WarehouseInventoryStats{
			Key:            itemKey,
			AvailableCount: stats.AvailableCount,
			InUseCount:     stats.InUseCount,
			TotalCount:     stats.TotalCount,
			TopHives:       stats.TopHives,
		}, nil
	}

	if strings.HasPrefix(itemKey, warehouseItemKeyPrefixFrameSpec) {
		specIDStr := strings.TrimPrefix(itemKey, warehouseItemKeyPrefixFrameSpec)
		specID, err := strconv.Atoi(specIDStr)
		if err != nil || specID <= 0 {
			return nil, fmt.Errorf("invalid frame spec key: %s", itemKey)
		}

		var availableCount int
		err = r.Db.Get(&availableCount, `
			SELECT count
			FROM warehouse_frame_inventory
			WHERE user_id=? AND frame_spec_id=?
			LIMIT 1
		`, r.UserID, specID)
		if err == sql.ErrNoRows {
			availableCount = 0
		} else if err != nil {
			return nil, err
		}

		var inUseCount int
		err = r.Db.Get(&inUseCount, `
			SELECT COUNT(*)
			FROM frames f
			INNER JOIN boxes b ON b.id = f.box_id AND b.user_id = f.user_id AND b.active = 1
			INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
			WHERE f.user_id = ?
			  AND f.active = 1
			  AND f.frame_spec_id = ?
			  AND h.active = 1
			  AND h.collapse_date IS NULL
			  AND h.merged_into_hive_id IS NULL
		`, r.UserID, specID)
		if err != nil {
			return nil, err
		}

		rows := []warehouseUsageScanRow{}
		err = r.Db.Select(&rows, `
			SELECT
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
			  AND f.frame_spec_id = ?
			  AND h.active = 1
			  AND h.collapse_date IS NULL
			  AND h.merged_into_hive_id IS NULL
			GROUP BY h.id, h.hive_number, h.apiary_id, a.name
			ORDER BY usage_count DESC, h.hive_number IS NULL ASC, h.hive_number ASC, h.id ASC
			LIMIT 10
		`, r.UserID, specID)
		if err != nil {
			return nil, err
		}

		topHives := mapUsageRows(rows)
		return &WarehouseInventoryStats{
			Key:            itemKey,
			AvailableCount: availableCount,
			InUseCount:     inUseCount,
			TotalCount:     availableCount + inUseCount,
			TopHives:       topHives,
		}, nil
	}

	return nil, fmt.Errorf("unsupported warehouse inventory key: %s", itemKey)
}

func parseBoxInventoryKey(itemKey string) (WarehouseModuleType, *int, error) {
	if !strings.HasPrefix(itemKey, warehouseItemKeyPrefixBox) {
		return "", nil, fmt.Errorf("invalid box inventory key: %s", itemKey)
	}
	payload := strings.TrimPrefix(itemKey, warehouseItemKeyPrefixBox)
	parts := strings.Split(payload, warehouseItemKeySystemDelimiter)
	if len(parts) > 2 {
		return "", nil, fmt.Errorf("invalid box inventory key: %s", itemKey)
	}

	moduleType := WarehouseModuleType(strings.TrimSpace(parts[0]))
	if !moduleType.IsValid() {
		return "", nil, fmt.Errorf("invalid box module type: %s", parts[0])
	}

	if len(parts) == 1 {
		return moduleType, nil, nil
	}

	systemID, err := strconv.Atoi(parts[1])
	if err != nil || systemID <= 0 {
		return "", nil, fmt.Errorf("invalid box system id in key: %s", itemKey)
	}
	return moduleType, &systemID, nil
}

func buildBoxInventoryKey(moduleType WarehouseModuleType, boxSystemID *int) string {
	if boxSystemID != nil && *boxSystemID > 0 {
		return warehouseItemKeyPrefixBox + moduleType.String() + warehouseItemKeySystemDelimiter + strconv.Itoa(*boxSystemID)
	}
	return warehouseItemKeyPrefixBox + moduleType.String()
}

func isLegacyFrameModuleType(moduleType WarehouseModuleType) bool {
	return moduleType == WarehouseModuleTypeFrameFoundation ||
		moduleType == WarehouseModuleTypeFrameEmptyComb ||
		moduleType == WarehouseModuleTypeFramePartition ||
		moduleType == WarehouseModuleTypeFrameFeeder
}

func isSystemScopedHivePartModuleType(moduleType WarehouseModuleType) bool {
	return moduleType == WarehouseModuleTypeDeep ||
		moduleType == WarehouseModuleTypeNucs ||
		moduleType == WarehouseModuleTypeSuper ||
		moduleType == WarehouseModuleTypeRoof ||
		moduleType == WarehouseModuleTypeHorizontalFeeder ||
		moduleType == WarehouseModuleTypeQueenExcluder ||
		moduleType == WarehouseModuleTypeBottom
}

func mapBoxModuleGroup(moduleType WarehouseModuleType) string {
	switch moduleType {
	case WarehouseModuleTypeDeep, WarehouseModuleTypeNucs, WarehouseModuleTypeSuper, WarehouseModuleTypeLargeHorizontalSection:
		return "HIVE_SECTIONS"
	default:
		return "HIVE_PARTS"
	}
}

func mapBoxModuleTitle(moduleType WarehouseModuleType) string {
	switch moduleType {
	case WarehouseModuleTypeDeep:
		return "Deep sections"
	case WarehouseModuleTypeNucs:
		return "Nucs"
	case WarehouseModuleTypeSuper:
		return "Super sections"
	case WarehouseModuleTypeLargeHorizontalSection:
		return "Horizontal hives"
	case WarehouseModuleTypeRoof:
		return "Roofs"
	case WarehouseModuleTypeHorizontalFeeder:
		return "Feeders"
	case WarehouseModuleTypeQueenExcluder:
		return "Queen excluders"
	case WarehouseModuleTypeBottom:
		return "Hive bottoms"
	default:
		return moduleType.String()
	}
}

func mapBoxModuleDescription(moduleType WarehouseModuleType) string {
	switch moduleType {
	case WarehouseModuleTypeDeep:
		return "Big hive sections used for brood and core colony space."
	case WarehouseModuleTypeNucs:
		return "Monolithic 5-frame nucleus hive bodies."
	case WarehouseModuleTypeSuper:
		return "Smaller sections usually used for honey storage."
	case WarehouseModuleTypeLargeHorizontalSection:
		return "Long horizontal hive bodies with high frame capacity."
	case WarehouseModuleTypeRoof:
		return "Top covers that protect the hive from rain and wind."
	case WarehouseModuleTypeHorizontalFeeder:
		return "Feeders used to provide syrup or supplements to colonies."
	case WarehouseModuleTypeQueenExcluder:
		return "Grids that keep the queen out of selected sections."
	case WarehouseModuleTypeBottom:
		return "Bottom boards used as the base of the hive."
	default:
		return ""
	}
}
