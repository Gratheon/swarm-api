-- +goose Up
CREATE TABLE IF NOT EXISTS `warehouse_settings` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(191) NOT NULL,
  `auto_update_from_hives` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_warehouse_settings_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `warehouse_settings`;
