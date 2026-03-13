-- +goose Up
ALTER TABLE `families`
  ADD COLUMN `active` tinyint(1) NOT NULL DEFAULT 1 AFTER `hive_id`;

ALTER TABLE `families`
  ADD INDEX `idx_families_user_hive_active` (`user_id`, `hive_id`, `active`);

-- +goose Down
ALTER TABLE `families`
  DROP INDEX `idx_families_user_hive_active`;

ALTER TABLE `families`
  DROP COLUMN `active`;
