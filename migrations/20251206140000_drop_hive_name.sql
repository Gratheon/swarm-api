-- +goose Up
-- First ensure all hive names have been copied to families
UPDATE families f
JOIN hives h ON h.family_id = f.id
SET f.name = h.name
WHERE h.active = 1 AND f.name IS NULL AND h.name IS NOT NULL;

-- Now we can safely drop the name column from hives
ALTER TABLE `hives` DROP COLUMN `name`;

-- +goose Down
-- Restore the name column (but data will be lost)
ALTER TABLE `hives` ADD COLUMN `name` VARCHAR(250) DEFAULT NULL;

