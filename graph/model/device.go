package model

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Device struct {
	Db *sqlx.DB

	ID        string     `json:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Name      string     `json:"name" db:"name"`
	Type      DeviceType `json:"type" db:"type"`
	APIToken  *string    `json:"apiToken" db:"api_token"`
	HiveID    *int       `json:"hiveId" db:"hive_id"`
	BoxID     *int       `json:"boxId" db:"box_id"`
	Active    bool       `json:"active" db:"active"`
	CreatedAt string     `json:"createdAt" db:"created_at"`
	UpdatedAt string     `json:"updatedAt" db:"updated_at"`
}

func isDevicesTableMissing(err error) bool {
	if err == nil {
		return false
	}

	mysqlErr, ok := err.(*mysqlDriver.MySQLError)
	if !ok {
		return false
	}

	return mysqlErr.Number == 1146 && strings.Contains(strings.ToLower(mysqlErr.Message), "devices")
}

func (r *Device) List() ([]*Device, error) {
	devices := []*Device{}
	err := r.Db.Select(&devices,
		`SELECT id, user_id, name, type, api_token, hive_id, box_id, active, created_at, updated_at
			FROM devices
			WHERE user_id=? AND active=1
		ORDER BY id DESC`,
		r.UserID,
	)
	if isDevicesTableMissing(err) {
		return []*Device{}, nil
	}
	return devices, err
}

func (r *Device) Get(id string) (*Device, error) {
	device := Device{}
	err := r.Db.Get(&device,
		`SELECT id, user_id, name, type, api_token, hive_id, box_id, active, created_at, updated_at
			FROM devices
			WHERE id=? AND user_id=? AND active=1
		LIMIT 1`,
		id, r.UserID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &device, err
}

func normalizeOptionalToken(token *string) *string {
	if token == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*token)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func (r *Device) resolveOptionalHiveID(hiveID *string) (*int, error) {
	if hiveID == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*hiveID)
	if trimmed == "" {
		return nil, nil
	}

	hiveIDInt, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil, err
	}

	var exists int
	err = r.Db.Get(&exists, "SELECT id FROM hives WHERE id=? AND user_id=? AND active=1 LIMIT 1", hiveIDInt, r.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("hive not found")
		}
		return nil, err
	}

	return &hiveIDInt, nil
}

func (r *Device) resolveOptionalBoxID(boxID *string) (*int, *int, error) {
	if boxID == nil {
		return nil, nil, nil
	}

	trimmed := strings.TrimSpace(*boxID)
	if trimmed == "" {
		return nil, nil, nil
	}

	boxIDInt, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil, nil, err
	}

	type boxRow struct {
		ID     int `db:"id"`
		HiveID int `db:"hive_id"`
	}

	box := boxRow{}
	err = r.Db.Get(&box, "SELECT id, hive_id FROM boxes WHERE id=? AND user_id=? AND active=1 LIMIT 1", boxIDInt, r.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("box not found")
		}
		return nil, nil, err
	}

	return &boxIDInt, &box.HiveID, nil
}

func (r *Device) resolveAssociationIDs(hiveID *string, boxID *string) (*int, *int, error) {
	resolvedHiveID, err := r.resolveOptionalHiveID(hiveID)
	if err != nil {
		return nil, nil, err
	}

	resolvedBoxID, boxHiveID, err := r.resolveOptionalBoxID(boxID)
	if err != nil {
		return nil, nil, err
	}

	if resolvedBoxID != nil {
		if resolvedHiveID != nil && *resolvedHiveID != *boxHiveID {
			return nil, nil, errors.New("selected box does not belong to selected hive")
		}
		resolvedHiveID = boxHiveID
	}

	return resolvedHiveID, resolvedBoxID, nil
}

func (r *Device) Create(input DeviceInput) (*Device, error) {
	tx := r.Db.MustBegin()

	apiToken := normalizeOptionalToken(input.APIToken)
	hiveID, boxID, err := r.resolveAssociationIDs(input.HiveID, input.BoxID)
	if err != nil {
		return nil, err
	}
	result, err := tx.NamedExec(
		`INSERT INTO devices (user_id, name, type, api_token, hive_id, box_id)
		VALUES (:userID, :name, :type, :apiToken, :hiveID, :boxID)`,
		map[string]interface{}{
			"userID":   r.UserID,
			"name":     strings.TrimSpace(input.Name),
			"type":     input.Type,
			"apiToken": apiToken,
			"hiveID":   hiveID,
			"boxID":    boxID,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.Get(stringID(id))
}

func (r *Device) Update(id string, input DeviceUpdateInput) (*Device, error) {
	current, err := r.Get(id)
	if err != nil || current == nil {
		return current, err
	}

	nextName := current.Name
	if input.Name != nil {
		nextName = strings.TrimSpace(*input.Name)
	}

	nextType := current.Type
	if input.Type != nil {
		nextType = *input.Type
	}

	nextToken := current.APIToken
	if input.APIToken != nil {
		nextToken = normalizeOptionalToken(input.APIToken)
	}

	nextHiveID := current.HiveID
	nextBoxID := current.BoxID
	if input.HiveID != nil || input.BoxID != nil {
		if input.HiveID != nil && input.BoxID == nil {
			resolvedHiveID, _, err := r.resolveAssociationIDs(input.HiveID, nil)
			if err != nil {
				return nil, err
			}
			nextHiveID = resolvedHiveID
			nextBoxID = nil
		} else {
			effectiveHiveID := input.HiveID
			effectiveBoxID := input.BoxID

			if effectiveHiveID == nil && current.HiveID != nil {
				hiveIDStr := strconv.Itoa(*current.HiveID)
				effectiveHiveID = &hiveIDStr
			}

			if effectiveBoxID == nil && current.BoxID != nil {
				boxIDStr := strconv.Itoa(*current.BoxID)
				effectiveBoxID = &boxIDStr
			}

			resolvedHiveID, resolvedBoxID, err := r.resolveAssociationIDs(effectiveHiveID, effectiveBoxID)
			if err != nil {
				return nil, err
			}
			nextHiveID = resolvedHiveID
			nextBoxID = resolvedBoxID
		}
	}

	tx := r.Db.MustBegin()
	_, err = tx.NamedExec(
		`UPDATE devices
		SET name=:name, type=:type, api_token=:apiToken, hive_id=:hiveID, box_id=:boxID
		WHERE id=:id AND user_id=:userID AND active=1`,
		map[string]interface{}{
			"id":       id,
			"userID":   r.UserID,
			"name":     nextName,
			"type":     nextType,
			"apiToken": nextToken,
			"hiveID":   nextHiveID,
			"boxID":    nextBoxID,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.Get(id)
}

func (r *Device) Deactivate(id string) (*bool, error) {
	success := true
	tx := r.Db.MustBegin()
	_, err := tx.NamedExec(
		"UPDATE devices SET active=0, api_token=NULL, hive_id=NULL, box_id=NULL WHERE id=:id AND user_id=:userID",
		map[string]interface{}{
			"id":     id,
			"userID": r.UserID,
		},
	)
	if err != nil {
		tx.Rollback()
		success = false
		return &success, err
	}

	if err := tx.Commit(); err != nil {
		success = false
		return &success, err
	}

	return &success, nil
}

func stringID(id int64) string {
	return strconv.FormatInt(id, 10)
}
