-- +goose Up
ALTER TABLE `families` ADD COLUMN `color` VARCHAR(10) DEFAULT NULL;

-- +goose Down
ALTER TABLE `families` DROP COLUMN `color`;

