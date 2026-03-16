-- +goose Up
ALTER TABLE `box_specs`
  ADD COLUMN `external_width_mm` int DEFAULT NULL AFTER `internal_height_mm`,
  ADD COLUMN `external_length_mm` int DEFAULT NULL AFTER `external_width_mm`;

-- +goose Down
ALTER TABLE `box_specs`
  DROP COLUMN `external_length_mm`,
  DROP COLUMN `external_width_mm`;
