-- +goose Up
CREATE TABLE IF NOT EXISTS `warehouse_modules` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(191) NOT NULL,
  `module_type` enum('DEEP','SUPER','ROOF','QUEEN_EXCLUDER','BOTTOM') NOT NULL,
  `count` int NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_module` (`user_id`, `module_type`),
  KEY `idx_warehouse_modules_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `warehouse_modules`;
