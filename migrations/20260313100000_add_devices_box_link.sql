-- +goose Up
ALTER TABLE `devices`
  ADD COLUMN `box_id` int unsigned DEFAULT NULL AFTER `hive_id`,
  ADD KEY `idx_devices_box_id` (`box_id`),
  ADD CONSTRAINT `fk_devices_box` FOREIGN KEY (`box_id`) REFERENCES `boxes` (`id`) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE `devices`
  DROP FOREIGN KEY `fk_devices_box`,
  DROP INDEX `idx_devices_box_id`,
  DROP COLUMN `box_id`;
