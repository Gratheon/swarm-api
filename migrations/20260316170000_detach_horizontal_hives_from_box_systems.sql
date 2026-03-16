-- +goose Up
UPDATE hives h
SET h.box_system_id = NULL
WHERE h.box_system_id IS NOT NULL
  AND EXISTS (
    SELECT 1
    FROM boxes b
    WHERE b.hive_id = h.id
      AND b.user_id = h.user_id
      AND b.active = 1
      AND b.type = 'LARGE_HORIZONTAL_SECTION'
  );

UPDATE boxes b
INNER JOIN hives h ON h.id = b.hive_id AND h.user_id = b.user_id
SET b.box_system_id = NULL
WHERE h.box_system_id IS NULL
  AND b.box_system_id IS NOT NULL;

-- +goose Down
-- no-op: previous box system linkage for horizontal hives cannot be reconstructed safely
SELECT 1;
