-- +goose Up
SET @has_hole_count := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'boxes'
    AND COLUMN_NAME = 'hole_count'
);
SET @sql := IF(
  @has_hole_count = 0,
  'ALTER TABLE `boxes` ADD COLUMN `hole_count` int NULL DEFAULT NULL AFTER `color`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +goose Down
ALTER TABLE `boxes` DROP COLUMN `hole_count`;
