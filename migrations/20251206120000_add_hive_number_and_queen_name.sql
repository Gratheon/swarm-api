-- +goose Up
ALTER TABLE `families` ADD COLUMN `name` VARCHAR(250) DEFAULT NULL;

ALTER TABLE `hives` ADD COLUMN `hive_number` INT DEFAULT NULL;

UPDATE families f
JOIN hives h ON h.family_id = f.id
SET f.name = h.name
WHERE h.active = 1 AND f.name IS NULL;

CREATE INDEX idx_hives_user_number ON hives(user_id, hive_number);

-- +goose Down
ALTER TABLE `families` DROP COLUMN `name`;
DROP INDEX idx_hives_user_number ON hives;
ALTER TABLE `hives` DROP COLUMN `hive_number`;

