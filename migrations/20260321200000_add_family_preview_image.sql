-- +goose Up
SET @has_preview_image_url := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'families'
    AND COLUMN_NAME = 'preview_image_url'
);
SET @sql := IF(@has_preview_image_url = 0,
  'ALTER TABLE `families` ADD COLUMN `preview_image_url` VARCHAR(255) DEFAULT NULL',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +goose Down
SET @has_preview_image_url := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'families'
    AND COLUMN_NAME = 'preview_image_url'
);
SET @sql := IF(@has_preview_image_url = 1,
  'ALTER TABLE `families` DROP COLUMN `preview_image_url`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
