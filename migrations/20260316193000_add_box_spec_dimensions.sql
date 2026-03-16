-- +goose Up
ALTER TABLE `box_specs`
  ADD COLUMN `internal_width_mm` int DEFAULT NULL AFTER `display_name`,
  ADD COLUMN `internal_length_mm` int DEFAULT NULL AFTER `internal_width_mm`,
  ADD COLUMN `internal_height_mm` int DEFAULT NULL AFTER `internal_length_mm`,
  ADD COLUMN `frame_width_mm` int DEFAULT NULL AFTER `internal_height_mm`,
  ADD COLUMN `frame_height_mm` int DEFAULT NULL AFTER `frame_width_mm`;

UPDATE `box_specs` bs
INNER JOIN `box_systems` sys ON sys.id = bs.system_id
SET
  bs.internal_width_mm = 375,
  bs.internal_length_mm = 465,
  bs.internal_height_mm = 244,
  bs.frame_width_mm = 482,
  bs.frame_height_mm = 232
WHERE sys.user_id IS NULL
  AND sys.name = 'Langstroth'
  AND bs.code = 'DEEP';

UPDATE `box_specs` bs
INNER JOIN `box_systems` sys ON sys.id = bs.system_id
SET
  bs.internal_width_mm = 375,
  bs.internal_length_mm = 465,
  bs.internal_height_mm = 168,
  bs.frame_width_mm = 482,
  bs.frame_height_mm = 159
WHERE sys.user_id IS NULL
  AND sys.name = 'Langstroth'
  AND bs.code = 'SUPER';

-- +goose Down
ALTER TABLE `box_specs`
  DROP COLUMN `frame_height_mm`,
  DROP COLUMN `frame_width_mm`,
  DROP COLUMN `internal_height_mm`,
  DROP COLUMN `internal_length_mm`,
  DROP COLUMN `internal_width_mm`;
