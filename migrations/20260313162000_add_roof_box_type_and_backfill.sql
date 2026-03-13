-- +goose Up
ALTER TABLE `boxes` CHANGE `type` `type`
ENUM(
  'SUPER',
  'DEEP',
  'ROOF',
  'LARGE_HORIZONTAL_SECTION',
  'GATE',
  'VENTILATION',
  'QUEEN_EXCLUDER',
  'HORIZONTAL_FEEDER',
  'BOTTOM'
) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'DEEP';

INSERT INTO `boxes` (`user_id`, `hive_id`, `active`, `color`, `position`, `type`)
SELECT
  h.`user_id`,
  h.`id`,
  1,
  '#363636',
  COALESCE((
    SELECT MAX(b.`position`) + 1
    FROM `boxes` b
    WHERE b.`hive_id` = h.`id`
      AND b.`user_id` = h.`user_id`
      AND b.`active` = 1
  ), 0),
  'ROOF'
FROM `hives` h
WHERE h.`active` = 1
  AND NOT EXISTS (
    SELECT 1
    FROM `boxes` rb
    WHERE rb.`hive_id` = h.`id`
      AND rb.`user_id` = h.`user_id`
      AND rb.`active` = 1
      AND rb.`type` = 'ROOF'
  );

-- +goose Down
UPDATE `boxes`
SET `active` = 0
WHERE `type` = 'ROOF' AND `active` = 1;

ALTER TABLE `boxes` CHANGE `type` `type`
ENUM(
  'SUPER',
  'DEEP',
  'LARGE_HORIZONTAL_SECTION',
  'GATE',
  'VENTILATION',
  'QUEEN_EXCLUDER',
  'HORIZONTAL_FEEDER',
  'BOTTOM'
) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'DEEP';
