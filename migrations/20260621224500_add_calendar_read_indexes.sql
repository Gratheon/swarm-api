-- +goose Up
ALTER TABLE `inspections` ADD INDEX `idx_calendar_inspections_user_added_hive` (`user_id`, `added`, `hive_id`);
ALTER TABLE `inspections` ADD INDEX `idx_calendar_inspections_user_hive_added` (`user_id`, `hive_id`, `added`);
ALTER TABLE `hive_logs` ADD INDEX `idx_calendar_hive_logs_user_created_hive` (`user_id`, `created_at`, `hive_id`);
ALTER TABLE `treatments` ADD INDEX `idx_calendar_treatments_user_added_hive_family` (`user_id`, `added`, `hive_id`, `family_id`);
ALTER TABLE `families` ADD INDEX `idx_calendar_families_user_hive_active_added` (`user_id`, `hive_id`, `active`, `added`);
ALTER TABLE `hives` ADD INDEX `idx_calendar_hives_user_active_apiary` (`user_id`, `active`, `apiary_id`);

-- +goose Down
ALTER TABLE `hives` DROP INDEX `idx_calendar_hives_user_active_apiary`;
ALTER TABLE `families` DROP INDEX `idx_calendar_families_user_hive_active_added`;
ALTER TABLE `treatments` DROP INDEX `idx_calendar_treatments_user_added_hive_family`;
ALTER TABLE `hive_logs` DROP INDEX `idx_calendar_hive_logs_user_created_hive`;
ALTER TABLE `inspections` DROP INDEX `idx_calendar_inspections_user_hive_added`;
ALTER TABLE `inspections` DROP INDEX `idx_calendar_inspections_user_added_hive`;
