-- +goose Up
CREATE TABLE `hive_placements` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL,
  `apiary_id` int unsigned NOT NULL,
  `hive_id` int unsigned NOT NULL,
  `x` float NOT NULL DEFAULT 0,
  `y` float NOT NULL DEFAULT 0,
  `rotation` float NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_hive_placement` (`apiary_id`, `hive_id`),
  KEY `idx_apiary_id` (`apiary_id`),
  KEY `idx_hive_id` (`hive_id`),
  CONSTRAINT `fk_hive_placement_apiary` FOREIGN KEY (`apiary_id`) REFERENCES `apiaries` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_hive_placement_hive` FOREIGN KEY (`hive_id`) REFERENCES `hives` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `apiary_obstacles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL,
  `apiary_id` int unsigned NOT NULL,
  `type` enum('CIRCLE','RECTANGLE') NOT NULL DEFAULT 'RECTANGLE',
  `x` float NOT NULL DEFAULT 0,
  `y` float NOT NULL DEFAULT 0,
  `width` float DEFAULT NULL,
  `height` float DEFAULT NULL,
  `radius` float DEFAULT NULL,
  `rotation` float NOT NULL DEFAULT 0,
  `label` varchar(100) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_apiary_id` (`apiary_id`),
  CONSTRAINT `fk_apiary_obstacle_apiary` FOREIGN KEY (`apiary_id`) REFERENCES `apiaries` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +goose Down
DROP TABLE IF EXISTS `apiary_obstacles`;
DROP TABLE IF EXISTS `hive_placements`;

