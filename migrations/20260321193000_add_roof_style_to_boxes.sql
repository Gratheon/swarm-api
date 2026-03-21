-- +goose Up
SET @has_roof_style := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'boxes'
    AND COLUMN_NAME = 'roof_style'
);
SET @sql := IF(@has_roof_style = 0,
  'ALTER TABLE `boxes` ADD COLUMN `roof_style` enum(''FLAT'',''ANGULAR'') NULL DEFAULT NULL AFTER `hole_count`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `boxes`
SET `roof_style` = 'FLAT'
WHERE `type` = 'ROOF' AND (`roof_style` IS NULL OR `roof_style` = '');

-- +goose Down
SET @has_roof_style := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'boxes'
    AND COLUMN_NAME = 'roof_style'
);
SET @sql := IF(@has_roof_style = 1,
  'ALTER TABLE `boxes` DROP COLUMN `roof_style`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
