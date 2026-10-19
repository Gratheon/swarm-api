package model

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
)

type HiveLog struct {
	Db *sqlx.DB

	ID           string                `json:"id" db:"id"`
	UserID       string                `db:"user_id"`
	HiveID       string                `json:"hiveId" db:"hive_id"`
	Action       string                `json:"action" db:"action"`
	Title        string                `json:"title" db:"title"`
	Details      *string               `json:"details" db:"details"`
	Source       *string               `json:"source" db:"source"`
	RelatedHives []*HiveLogRelatedHive `json:"relatedHives" db:"-"`
	DedupeKey    *string               `json:"dedupeKey" db:"dedupe_key"`
	CreatedAt    string                `json:"createdAt" db:"created_at"`
	UpdatedAt    string                `json:"updatedAt" db:"updated_at"`

	RelatedHivesRaw *string `db:"related_hives"`
}

type hiveLogRow struct {
	ID              string  `db:"id"`
	UserID          string  `db:"user_id"`
	HiveID          string  `db:"hive_id"`
	Action          string  `db:"action"`
	Title           string  `db:"title"`
	Details         *string `db:"details"`
	Source          *string `db:"source"`
	RelatedHivesRaw *string `db:"related_hives"`
	DedupeKey       *string `db:"dedupe_key"`
	CreatedAt       string  `db:"created_at"`
	UpdatedAt       string  `db:"updated_at"`
}

func (r *HiveLog) ensureHiveOwnership(hiveID string) error {
	var count int
	err := r.Db.Get(&count,
		`SELECT COUNT(*) FROM hives WHERE id=$1 AND user_id=$2 AND active=TRUE`,
		hiveID,
		r.UserID,
	)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("hive not found")
	}
	return nil
}

func (r *HiveLog) ListByHive(hiveID string, limit *int) ([]*HiveLog, error) {
	if err := r.ensureHiveOwnership(hiveID); err != nil {
		return nil, err
	}

	finalLimit := 200
	if limit != nil && *limit > 0 {
		finalLimit = *limit
		if finalLimit > 1000 {
			finalLimit = 1000
		}
	}

	rows := []*hiveLogRow{}
	err := r.Db.Select(&rows,
		`SELECT id, user_id, hive_id, action, title, details, source, related_hives, dedupe_key, created_at, updated_at
		 FROM hive_logs
		 WHERE user_id=$1 AND hive_id=$2 AND active=TRUE
		 ORDER BY created_at DESC, id DESC
		 LIMIT $3`,
		r.UserID,
		hiveID,
		finalLimit,
	)
	if err != nil {
		return nil, err
	}

	result := make([]*HiveLog, 0, len(rows))
	for _, row := range rows {
		log := &HiveLog{
			ID:           row.ID,
			UserID:       row.UserID,
			HiveID:       row.HiveID,
			Action:       row.Action,
			Title:        row.Title,
			Details:      row.Details,
			Source:       row.Source,
			DedupeKey:    row.DedupeKey,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
			RelatedHives: []*HiveLogRelatedHive{},
		}
		if row.RelatedHivesRaw != nil && *row.RelatedHivesRaw != "" {
			_ = json.Unmarshal([]byte(*row.RelatedHivesRaw), &log.RelatedHives)
		}
		result = append(result, log)
	}

	return result, nil
}

func (r *HiveLog) Create(input HiveLogInput) (*HiveLog, error) {
	if err := r.ensureHiveOwnership(input.HiveID); err != nil {
		return nil, err
	}

	relatedBytes, _ := json.Marshal(input.RelatedHives)
	relatedJSON := string(relatedBytes)

	// A repeated dedupe_key returns the existing row (and bumps updated_at) instead of a duplicate
	id, err := namedInsertReturningID(r.Db,
		`INSERT INTO hive_logs (user_id, hive_id, action, title, details, source, related_hives, dedupe_key)
		 VALUES (:user_id, :hive_id, :action, :title, :details, :source, :related_hives, :dedupe_key)
		 ON CONFLICT (user_id, dedupe_key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		 RETURNING id`,
		map[string]interface{}{
			"user_id":       r.UserID,
			"hive_id":       input.HiveID,
			"action":        input.Action,
			"title":         input.Title,
			"details":       input.Details,
			"source":        input.Source,
			"related_hives": relatedJSON,
			"dedupe_key":    input.DedupeKey,
		},
	)
	if err != nil {
		return nil, err
	}

	row := &hiveLogRow{}
	err = r.Db.Get(row,
		`SELECT id, user_id, hive_id, action, title, details, source, related_hives, dedupe_key, created_at, updated_at
		 FROM hive_logs
		 WHERE id=$1 AND user_id=$2 AND active=TRUE LIMIT 1`,
		id,
		r.UserID,
	)
	if err != nil {
		return nil, err
	}

	result := &HiveLog{
		ID:           row.ID,
		UserID:       row.UserID,
		HiveID:       row.HiveID,
		Action:       row.Action,
		Title:        row.Title,
		Details:      row.Details,
		Source:       row.Source,
		DedupeKey:    row.DedupeKey,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		RelatedHives: []*HiveLogRelatedHive{},
	}
	if row.RelatedHivesRaw != nil && *row.RelatedHivesRaw != "" {
		_ = json.Unmarshal([]byte(*row.RelatedHivesRaw), &result.RelatedHives)
	}

	return result, nil
}

func (r *HiveLog) Update(id string, input HiveLogUpdateInput) (*HiveLog, error) {
	var ownerHiveID string
	err := r.Db.Get(&ownerHiveID,
		`SELECT hive_id FROM hive_logs WHERE id=$1 AND user_id=$2 AND active=TRUE LIMIT 1`,
		id,
		r.UserID,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("hive log not found")
	}
	if err != nil {
		return nil, err
	}

	if err := r.ensureHiveOwnership(ownerHiveID); err != nil {
		return nil, err
	}

	relatedBytes, _ := json.Marshal(input.RelatedHives)
	relatedJSON := string(relatedBytes)

	_, err = r.Db.NamedExec(
		`UPDATE hive_logs
		 SET title = COALESCE(:title, title),
		     details = COALESCE(:details, details),
		     related_hives = CASE WHEN :has_related THEN :related_hives ELSE related_hives END,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id=:id AND user_id=:user_id AND active=TRUE`,
		map[string]interface{}{
			"id":            id,
			"user_id":       r.UserID,
			"title":         input.Title,
			"details":       input.Details,
			"has_related":   input.RelatedHives != nil,
			"related_hives": relatedJSON,
		},
	)
	if err != nil {
		return nil, err
	}

	row := &hiveLogRow{}
	err = r.Db.Get(row,
		`SELECT id, user_id, hive_id, action, title, details, source, related_hives, dedupe_key, created_at, updated_at
		 FROM hive_logs
		 WHERE id=$1 AND user_id=$2 AND active=TRUE LIMIT 1`,
		id,
		r.UserID,
	)
	if err != nil {
		return nil, err
	}

	result := &HiveLog{
		ID:           row.ID,
		UserID:       row.UserID,
		HiveID:       row.HiveID,
		Action:       row.Action,
		Title:        row.Title,
		Details:      row.Details,
		Source:       row.Source,
		DedupeKey:    row.DedupeKey,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		RelatedHives: []*HiveLogRelatedHive{},
	}
	if row.RelatedHivesRaw != nil && *row.RelatedHivesRaw != "" {
		_ = json.Unmarshal([]byte(*row.RelatedHivesRaw), &result.RelatedHives)
	}

	return result, nil
}

func (r *HiveLog) Delete(id string) (bool, error) {
	result, err := r.Db.Exec(
		`UPDATE hive_logs SET active=FALSE, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND user_id=$2 AND active=TRUE`,
		id,
		r.UserID,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
