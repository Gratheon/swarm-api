-- +goose Up
ALTER TABLE `warehouse_modules`
  ADD COLUMN `box_system_id` int NOT NULL DEFAULT 0 AFTER `module_type`;

ALTER TABLE `warehouse_modules`
  DROP INDEX `uniq_user_module`,
  ADD UNIQUE KEY `uniq_user_module_system` (`user_id`, `module_type`, `box_system_id`),
  ADD KEY `idx_warehouse_modules_user_system` (`user_id`, `box_system_id`);

-- +goose Down
ALTER TABLE `warehouse_modules`
  DROP INDEX `uniq_user_module_system`,
  DROP INDEX `idx_warehouse_modules_user_system`,
  ADD UNIQUE KEY `uniq_user_module` (`user_id`, `module_type`);

ALTER TABLE `warehouse_modules`
  DROP COLUMN `box_system_id`;
