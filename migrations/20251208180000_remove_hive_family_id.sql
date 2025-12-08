-- +goose Up
-- Remove redundant hive.family_id column since we now use family.hive_id
-- This simplifies the data model to have only one direction: family -> hive

-- First, ensure all families.hive_id are synced from hives.family_id (for any data that wasn't migrated)
UPDATE families f
INNER JOIN hives h ON h.family_id = f.id
SET f.hive_id = h.id
WHERE f.hive_id IS NULL
  AND h.active = 1;

-- Now drop the redundant column
ALTER TABLE hives DROP COLUMN family_id;

-- +goose Down
-- Restore the column
ALTER TABLE hives ADD COLUMN family_id INT DEFAULT NULL AFTER apiary_id;

-- Repopulate from families.hive_id
UPDATE hives h
INNER JOIN (
    SELECT hive_id, MIN(id) as family_id
    FROM families
    WHERE hive_id IS NOT NULL
    GROUP BY hive_id
) f ON h.id = f.hive_id
SET h.family_id = f.family_id;

