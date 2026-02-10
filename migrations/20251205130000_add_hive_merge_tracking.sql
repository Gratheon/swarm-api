-- +goose Up
ALTER TABLE hives
  ADD COLUMN merged_into_hive_id INT UNSIGNED NULL,
  ADD COLUMN merge_date DATETIME NULL,
  ADD COLUMN merge_type ENUM('both_queens', 'source_queen_kept', 'target_queen_kept') NULL,
  ADD INDEX idx_merged_into_hive_id (merged_into_hive_id);

-- +goose Down
ALTER TABLE hives
  DROP INDEX idx_merged_into_hive_id,
  DROP COLUMN merge_type,
  DROP COLUMN merge_date,
  DROP COLUMN merged_into_hive_id;

