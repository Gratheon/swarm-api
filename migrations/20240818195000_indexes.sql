-- +goose Up
ALTER TABLE `frames_sides` ADD INDEX (`user_id`);
ALTER TABLE `frames` ADD INDEX (`user_id`, `box_id`);
ALTER TABLE `apiaries` ADD INDEX (`user_id`, `active`);
ALTER TABLE `boxes` ADD INDEX (`user_id`, `hive_id`, `active`);
ALTER TABLE `hives` ADD INDEX (`user_id`, `apiary_id`, `active`);
ALTER TABLE `inspections` ADD INDEX (`user_id`, `hive_id`);
