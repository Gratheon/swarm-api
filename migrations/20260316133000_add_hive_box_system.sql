-- +goose Up
SET @has_hive_box_system_id := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'hives'
    AND COLUMN_NAME = 'box_system_id'
);
SET @sql := IF(@has_hive_box_system_id = 0,
  'ALTER TABLE `hives` ADD COLUMN `box_system_id` int DEFAULT NULL AFTER `apiary_id`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_idx_hives_box_system := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'hives'
    AND INDEX_NAME = 'idx_hives_box_system'
);
SET @sql := IF(@has_idx_hives_box_system = 0,
  'CREATE INDEX `idx_hives_box_system` ON `hives` (`user_id`, `box_system_id`)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `hives` h
LEFT JOIN `boxes` b
  ON b.hive_id = h.id
 AND b.user_id = h.user_id
 AND b.active = 1
LEFT JOIN `box_systems` bs
  ON bs.id = b.box_system_id
SET h.box_system_id = COALESCE(b.box_system_id, (
  SELECT id
  FROM box_systems
  WHERE user_id IS NULL
    AND name = 'Langstroth'
    AND active = 1
  ORDER BY is_default DESC, id ASC
  LIMIT 1
))
WHERE h.box_system_id IS NULL;

-- +goose Down
ALTER TABLE `hives`
  DROP INDEX `idx_hives_box_system`,
  DROP COLUMN `box_system_id`;
