-- +goose Up
ALTER TABLE `box_systems`
  ADD COLUMN `box_profile_source_system_id` int DEFAULT NULL AFTER `name`,
  ADD KEY `idx_box_systems_box_profile_source` (`box_profile_source_system_id`),
  ADD CONSTRAINT `fk_box_systems_box_profile_source`
    FOREIGN KEY (`box_profile_source_system_id`) REFERENCES `box_systems`(`id`) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE `box_systems`
  DROP FOREIGN KEY `fk_box_systems_box_profile_source`,
  DROP INDEX `idx_box_systems_box_profile_source`,
  DROP COLUMN `box_profile_source_system_id`;
