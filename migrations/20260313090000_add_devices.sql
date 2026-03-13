-- +goose Up
CREATE TABLE `devices` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL,
  `name` varchar(250) NOT NULL,
  `type` enum('IOT_SENSOR','VIDEO_CAMERA') NOT NULL,
  `api_token` varchar(512) DEFAULT NULL,
  `hive_id` int unsigned DEFAULT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_device_api_token` (`api_token`),
  KEY `idx_devices_user_active` (`user_id`, `active`),
  KEY `idx_devices_hive_id` (`hive_id`),
  CONSTRAINT `fk_devices_hive` FOREIGN KEY (`hive_id`) REFERENCES `hives` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +goose Down
DROP TABLE IF EXISTS `devices`;
