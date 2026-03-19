-- +goose Up
ALTER TABLE `warehouse_modules`
MODIFY COLUMN `module_type` enum(
  'DEEP',
  'NUCS',
  'SUPER',
  'LARGE_HORIZONTAL_SECTION',
  'ROOF',
  'HORIZONTAL_FEEDER',
  'QUEEN_EXCLUDER',
  'BOTTOM',
  'FRAME_FOUNDATION',
  'FRAME_EMPTY_COMB',
  'FRAME_PARTITION',
  'FRAME_FEEDER'
) NOT NULL;

-- +goose Down
UPDATE warehouse_modules
SET module_type = 'DEEP'
WHERE module_type = 'NUCS';

ALTER TABLE `warehouse_modules`
MODIFY COLUMN `module_type` enum(
  'DEEP',
  'SUPER',
  'LARGE_HORIZONTAL_SECTION',
  'ROOF',
  'HORIZONTAL_FEEDER',
  'QUEEN_EXCLUDER',
  'BOTTOM',
  'FRAME_FOUNDATION',
  'FRAME_EMPTY_COMB',
  'FRAME_PARTITION',
  'FRAME_FEEDER'
) NOT NULL;
