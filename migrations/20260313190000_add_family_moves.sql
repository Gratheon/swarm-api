-- +goose Up
CREATE TABLE IF NOT EXISTS `family_moves` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL,
  `family_id` int unsigned NOT NULL,
  `from_hive_id` int unsigned DEFAULT NULL,
  `to_hive_id` int unsigned DEFAULT NULL,
  `move_type` varchar(32) NOT NULL DEFAULT 'MOVED',
  `moved_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_family_moves_user_family_time` (`user_id`, `family_id`, `moved_at`),
  KEY `idx_family_moves_from_hive` (`from_hive_id`),
  KEY `idx_family_moves_to_hive` (`to_hive_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `family_moves` (`user_id`, `family_id`, `from_hive_id`, `to_hive_id`, `move_type`)
SELECT f.`user_id`, f.`id`, NULL, f.`hive_id`, 'BACKFILL_ASSIGNED'
FROM `families` f
WHERE f.`hive_id` IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS `family_moves`;
