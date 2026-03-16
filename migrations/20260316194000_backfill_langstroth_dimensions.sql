-- +goose Up
UPDATE `box_specs` bs
INNER JOIN `box_systems` sys ON sys.id = bs.system_id
SET
  bs.internal_width_mm = COALESCE(bs.internal_width_mm, 375),
  bs.internal_length_mm = COALESCE(bs.internal_length_mm, 465),
  bs.internal_height_mm = COALESCE(bs.internal_height_mm, 244),
  bs.frame_width_mm = COALESCE(bs.frame_width_mm, 482),
  bs.frame_height_mm = COALESCE(bs.frame_height_mm, 232)
WHERE LOWER(TRIM(sys.name)) = 'langstroth'
  AND bs.code = 'DEEP'
  AND (
    bs.internal_width_mm IS NULL
    OR bs.internal_length_mm IS NULL
    OR bs.internal_height_mm IS NULL
    OR bs.frame_width_mm IS NULL
    OR bs.frame_height_mm IS NULL
  );

UPDATE `box_specs` bs
INNER JOIN `box_systems` sys ON sys.id = bs.system_id
SET
  bs.internal_width_mm = COALESCE(bs.internal_width_mm, 375),
  bs.internal_length_mm = COALESCE(bs.internal_length_mm, 465),
  bs.internal_height_mm = COALESCE(bs.internal_height_mm, 168),
  bs.frame_width_mm = COALESCE(bs.frame_width_mm, 482),
  bs.frame_height_mm = COALESCE(bs.frame_height_mm, 159)
WHERE LOWER(TRIM(sys.name)) = 'langstroth'
  AND bs.code = 'SUPER'
  AND (
    bs.internal_width_mm IS NULL
    OR bs.internal_length_mm IS NULL
    OR bs.internal_height_mm IS NULL
    OR bs.frame_width_mm IS NULL
    OR bs.frame_height_mm IS NULL
  );

-- +goose Down
-- no-op: avoid wiping potentially user-edited dimensions
SELECT 1;
