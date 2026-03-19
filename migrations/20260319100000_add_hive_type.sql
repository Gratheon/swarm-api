-- +goose Up
SET @has_hive_type := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'hives'
    AND COLUMN_NAME = 'hive_type'
);
SET @sql := IF(
  @has_hive_type = 0,
  'ALTER TABLE `hives` ADD COLUMN `hive_type` varchar(32) NOT NULL DEFAULT ''VERTICAL'' AFTER `box_system_id`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE hives h
SET h.hive_type = CASE
  WHEN EXISTS (
    SELECT 1
    FROM boxes b
    WHERE b.hive_id = h.id
      AND b.active = 1
      AND b.type = 'LARGE_HORIZONTAL_SECTION'
  ) THEN 'HORIZONTAL'
  WHEN h.hive_type IS NULL OR h.hive_type = '' THEN 'VERTICAL'
  ELSE h.hive_type
END
WHERE h.active = 1;

UPDATE hives
SET hive_type = 'VERTICAL'
WHERE hive_type IS NULL OR hive_type = '';

-- +goose Down
ALTER TABLE `hives` DROP COLUMN `hive_type`;
