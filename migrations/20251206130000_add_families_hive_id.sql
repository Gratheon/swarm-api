-- +goose Up
ALTER TABLE `families` ADD COLUMN `hive_id` INT DEFAULT NULL;

UPDATE families f
JOIN hives h ON h.family_id = f.id
SET f.hive_id = h.id
WHERE h.active = 1 AND f.hive_id IS NULL;

CREATE INDEX idx_families_hive ON families(hive_id);

-- +goose Down
DROP INDEX idx_families_hive ON families;
ALTER TABLE `families` DROP COLUMN `hive_id`;

