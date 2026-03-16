-- +goose Up
CREATE TABLE IF NOT EXISTS `box_systems` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(191) DEFAULT NULL,
  `name` varchar(191) NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_box_systems_user` (`user_id`),
  KEY `idx_box_systems_active` (`active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `box_specs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `system_id` int NOT NULL,
  `code` varchar(64) NOT NULL,
  `legacy_box_type` enum(
    'SUPER',
    'DEEP',
    'LARGE_HORIZONTAL_SECTION',
    'GATE',
    'VENTILATION',
    'QUEEN_EXCLUDER',
    'HORIZONTAL_FEEDER',
    'BOTTOM',
    'ROOF'
  ) NOT NULL,
  `display_name` varchar(191) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_box_specs_system_code` (`system_id`, `code`),
  UNIQUE KEY `uniq_box_specs_system_legacy` (`system_id`, `legacy_box_type`),
  KEY `idx_box_specs_system` (`system_id`),
  CONSTRAINT `fk_box_specs_system` FOREIGN KEY (`system_id`) REFERENCES `box_systems`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `frame_specs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `system_id` int NOT NULL,
  `code` varchar(64) NOT NULL,
  `frame_type` enum('VOID','FOUNDATION','EMPTY_COMB','PARTITION','FEEDER') NOT NULL,
  `display_name` varchar(191) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_frame_specs_system_code` (`system_id`, `code`),
  KEY `idx_frame_specs_system_type` (`system_id`, `frame_type`),
  CONSTRAINT `fk_frame_specs_system` FOREIGN KEY (`system_id`) REFERENCES `box_systems`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `frame_spec_compatibility` (
  `id` int NOT NULL AUTO_INCREMENT,
  `frame_spec_id` int NOT NULL,
  `box_spec_id` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_frame_box_compat` (`frame_spec_id`, `box_spec_id`),
  KEY `idx_frame_spec_compat_box` (`box_spec_id`),
  CONSTRAINT `fk_frame_spec_compat_frame` FOREIGN KEY (`frame_spec_id`) REFERENCES `frame_specs`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_frame_spec_compat_box` FOREIGN KEY (`box_spec_id`) REFERENCES `box_specs`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET @has_box_system_id := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'boxes'
    AND COLUMN_NAME = 'box_system_id'
);
SET @sql := IF(@has_box_system_id = 0,
  'ALTER TABLE `boxes` ADD COLUMN `box_system_id` int DEFAULT NULL AFTER `type`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_box_spec_id := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'boxes'
    AND COLUMN_NAME = 'box_spec_id'
);
SET @sql := IF(@has_box_spec_id = 0,
  'ALTER TABLE `boxes` ADD COLUMN `box_spec_id` int DEFAULT NULL AFTER `box_system_id`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_idx_boxes_system_spec := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'boxes'
    AND INDEX_NAME = 'idx_boxes_system_spec'
);
SET @sql := IF(@has_idx_boxes_system_spec = 0,
  'CREATE INDEX `idx_boxes_system_spec` ON `boxes` (`user_id`, `box_system_id`, `box_spec_id`)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_frame_spec_id := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'frames'
    AND COLUMN_NAME = 'frame_spec_id'
);
SET @sql := IF(@has_frame_spec_id = 0,
  'ALTER TABLE `frames` ADD COLUMN `frame_spec_id` int DEFAULT NULL AFTER `type`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_idx_frames_spec := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'frames'
    AND INDEX_NAME = 'idx_frames_spec'
);
SET @sql := IF(@has_idx_frames_spec = 0,
  'CREATE INDEX `idx_frames_spec` ON `frames` (`user_id`, `frame_spec_id`)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `box_systems` (`user_id`, `name`, `is_default`, `active`)
SELECT NULL, 'Langstroth', 1, 1
WHERE NOT EXISTS (
  SELECT 1 FROM `box_systems` WHERE `user_id` IS NULL AND `name` = 'Langstroth' AND `active` = 1
);

INSERT INTO `box_specs` (`system_id`, `code`, `legacy_box_type`, `display_name`)
SELECT bs.id, x.code, x.legacy_box_type, x.display_name
FROM (
  SELECT 'DEEP' AS code, 'DEEP' AS legacy_box_type, 'Deep' AS display_name
  UNION ALL SELECT 'SUPER', 'SUPER', 'Super'
  UNION ALL SELECT 'HORIZONTAL', 'LARGE_HORIZONTAL_SECTION', 'Horizontal section'
  UNION ALL SELECT 'ROOF', 'ROOF', 'Roof'
  UNION ALL SELECT 'GATE', 'GATE', 'Entrance gate'
  UNION ALL SELECT 'VENTILATION', 'VENTILATION', 'Ventilation'
  UNION ALL SELECT 'QUEEN_EXCLUDER', 'QUEEN_EXCLUDER', 'Queen excluder'
  UNION ALL SELECT 'HORIZONTAL_FEEDER', 'HORIZONTAL_FEEDER', 'Horizontal feeder'
  UNION ALL SELECT 'BOTTOM', 'BOTTOM', 'Bottom'
) x
JOIN `box_systems` bs
  ON bs.user_id IS NULL
 AND bs.name = 'Langstroth'
LEFT JOIN `box_specs` existing
  ON existing.system_id = bs.id
 AND existing.code = x.code
WHERE existing.id IS NULL;

INSERT INTO `frame_specs` (`system_id`, `code`, `frame_type`, `display_name`)
SELECT bs.id, x.code, x.frame_type, x.display_name
FROM (
  SELECT 'FOUNDATION_DEEP' AS code, 'FOUNDATION' AS frame_type, 'Foundation (deep)' AS display_name
  UNION ALL SELECT 'FOUNDATION_SUPER', 'FOUNDATION', 'Foundation (super)'
  UNION ALL SELECT 'FOUNDATION_HORIZONTAL', 'FOUNDATION', 'Foundation (horizontal)'
  UNION ALL SELECT 'VOID_DEEP', 'VOID', 'Empty frame (deep)'
  UNION ALL SELECT 'VOID_SUPER', 'VOID', 'Empty frame (super)'
  UNION ALL SELECT 'VOID_HORIZONTAL', 'VOID', 'Empty frame (horizontal)'
  UNION ALL SELECT 'EMPTY_COMB_DEEP', 'EMPTY_COMB', 'Empty comb (deep)'
  UNION ALL SELECT 'EMPTY_COMB_SUPER', 'EMPTY_COMB', 'Empty comb (super)'
  UNION ALL SELECT 'EMPTY_COMB_HORIZONTAL', 'EMPTY_COMB', 'Empty comb (horizontal)'
  UNION ALL SELECT 'PARTITION_DEEP', 'PARTITION', 'Partition (deep)'
  UNION ALL SELECT 'PARTITION_SUPER', 'PARTITION', 'Partition (super)'
  UNION ALL SELECT 'PARTITION_HORIZONTAL', 'PARTITION', 'Partition (horizontal)'
  UNION ALL SELECT 'FEEDER_DEEP', 'FEEDER', 'Frame feeder (deep)'
  UNION ALL SELECT 'FEEDER_SUPER', 'FEEDER', 'Frame feeder (super)'
  UNION ALL SELECT 'FEEDER_HORIZONTAL', 'FEEDER', 'Frame feeder (horizontal)'
) x
JOIN `box_systems` bs
  ON bs.user_id IS NULL
 AND bs.name = 'Langstroth'
LEFT JOIN `frame_specs` existing
  ON existing.system_id = bs.id
 AND existing.code = x.code
WHERE existing.id IS NULL;

INSERT INTO `frame_spec_compatibility` (`frame_spec_id`, `box_spec_id`)
SELECT fs.id, bs.id
FROM `frame_specs` fs
JOIN `box_systems` sys ON sys.id = fs.system_id AND sys.user_id IS NULL AND sys.name = 'Langstroth'
JOIN `box_specs` bs ON bs.system_id = sys.id
LEFT JOIN `frame_spec_compatibility` existing
  ON existing.frame_spec_id = fs.id
 AND existing.box_spec_id = bs.id
WHERE existing.id IS NULL
  AND (
    (fs.code LIKE '%_DEEP' AND bs.code = 'DEEP') OR
    (fs.code LIKE '%_SUPER' AND bs.code = 'SUPER') OR
    (fs.code LIKE '%_HORIZONTAL' AND bs.code = 'HORIZONTAL')
  );

UPDATE `boxes` b
JOIN `box_specs` bs ON bs.legacy_box_type = (CONVERT(b.type USING utf8mb4) COLLATE utf8mb4_unicode_ci)
JOIN `box_systems` sys ON sys.id = bs.system_id AND sys.user_id IS NULL AND sys.name = 'Langstroth'
SET b.box_system_id = sys.id,
    b.box_spec_id = bs.id
WHERE b.box_system_id IS NULL OR b.box_spec_id IS NULL;

UPDATE `frames` f
JOIN `boxes` b
  ON b.id = f.box_id
 AND b.user_id = f.user_id
 AND b.active = 1
JOIN `frame_specs` fs
  ON fs.system_id = b.box_system_id
 AND fs.frame_type = (CONVERT(f.type USING utf8mb4) COLLATE utf8mb4_unicode_ci)
JOIN `frame_spec_compatibility` c
  ON c.frame_spec_id = fs.id
 AND c.box_spec_id = b.box_spec_id
SET f.frame_spec_id = fs.id
WHERE f.active = 1;

-- +goose Down
ALTER TABLE `frames` DROP INDEX `idx_frames_spec`, DROP COLUMN `frame_spec_id`;
ALTER TABLE `boxes` DROP INDEX `idx_boxes_system_spec`, DROP COLUMN `box_spec_id`, DROP COLUMN `box_system_id`;

DROP TABLE IF EXISTS `frame_spec_compatibility`;
DROP TABLE IF EXISTS `frame_specs`;
DROP TABLE IF EXISTS `box_specs`;
DROP TABLE IF EXISTS `box_systems`;
