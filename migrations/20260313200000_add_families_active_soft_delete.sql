-- +goose Up
SET @active_col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'families'
    AND COLUMN_NAME = 'active'
);

SET @add_active_col_sql := IF(
  @active_col_exists = 0,
  'ALTER TABLE `families` ADD COLUMN `active` tinyint(1) NOT NULL DEFAULT 1 AFTER `hive_id`',
  'SELECT 1'
);
PREPARE add_active_col_stmt FROM @add_active_col_sql;
EXECUTE add_active_col_stmt;
DEALLOCATE PREPARE add_active_col_stmt;

SET @active_idx_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'families'
    AND INDEX_NAME = 'idx_families_user_hive_active'
);

SET @add_active_idx_sql := IF(
  @active_idx_exists = 0,
  'ALTER TABLE `families` ADD INDEX `idx_families_user_hive_active` (`user_id`, `hive_id`, `active`)',
  'SELECT 1'
);
PREPARE add_active_idx_stmt FROM @add_active_idx_sql;
EXECUTE add_active_idx_stmt;
DEALLOCATE PREPARE add_active_idx_stmt;

-- +goose Down
SET @active_idx_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'families'
    AND INDEX_NAME = 'idx_families_user_hive_active'
);

SET @drop_active_idx_sql := IF(
  @active_idx_exists > 0,
  'ALTER TABLE `families` DROP INDEX `idx_families_user_hive_active`',
  'SELECT 1'
);
PREPARE drop_active_idx_stmt FROM @drop_active_idx_sql;
EXECUTE drop_active_idx_stmt;
DEALLOCATE PREPARE drop_active_idx_stmt;

SET @active_col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'families'
    AND COLUMN_NAME = 'active'
);

SET @drop_active_col_sql := IF(
  @active_col_exists > 0,
  'ALTER TABLE `families` DROP COLUMN `active`',
  'SELECT 1'
);
PREPARE drop_active_col_stmt FROM @drop_active_col_sql;
EXECUTE drop_active_col_stmt;
DEALLOCATE PREPARE drop_active_col_stmt;
