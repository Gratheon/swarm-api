-- +goose Up
CREATE TABLE IF NOT EXISTS `warehouse_frame_inventory` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(191) NOT NULL,
  `frame_spec_id` int NOT NULL,
  `count` int NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_warehouse_frame_inventory` (`user_id`, `frame_spec_id`),
  KEY `idx_warehouse_frame_inventory_user` (`user_id`),
  KEY `idx_warehouse_frame_inventory_spec` (`frame_spec_id`),
  CONSTRAINT `fk_warehouse_frame_inventory_spec` FOREIGN KEY (`frame_spec_id`) REFERENCES `frame_specs`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS `warehouse_frame_inventory`;
