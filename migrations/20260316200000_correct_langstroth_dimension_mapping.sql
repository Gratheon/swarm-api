-- +goose Up
UPDATE `box_specs` bs
INNER JOIN `box_systems` sys ON sys.id = bs.system_id
SET
  bs.internal_width_mm = 465,
  bs.internal_length_mm = 375,
  bs.frame_width_mm = 448
WHERE LOWER(TRIM(sys.name)) = 'langstroth'
  AND bs.code = 'DEEP'
  AND (
    bs.internal_width_mm IS NULL
    OR bs.internal_length_mm IS NULL
    OR bs.frame_width_mm IS NULL
    OR bs.internal_width_mm = 375
    OR bs.internal_length_mm = 465
    OR bs.frame_width_mm = 482
  );

UPDATE `box_specs` bs
INNER JOIN `box_systems` sys ON sys.id = bs.system_id
SET
  bs.internal_width_mm = 465,
  bs.internal_length_mm = 375,
  bs.frame_width_mm = 448
WHERE LOWER(TRIM(sys.name)) = 'langstroth'
  AND bs.code = 'SUPER'
  AND (
    bs.internal_width_mm IS NULL
    OR bs.internal_length_mm IS NULL
    OR bs.frame_width_mm IS NULL
    OR bs.internal_width_mm = 375
    OR bs.internal_length_mm = 465
    OR bs.frame_width_mm = 482
  );

-- +goose Down
-- no-op: do not overwrite potentially edited dimensions
SELECT 1;
