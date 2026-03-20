-- +goose Up
UPDATE frame_specs
SET active = 1
WHERE frame_type IN ('PARTITION', 'FEEDER');

INSERT INTO frame_spec_compatibility (frame_spec_id, box_spec_id)
SELECT fs.id, bs.id
FROM frame_specs fs
INNER JOIN box_specs bs ON bs.system_id = fs.system_id AND bs.active = 1
LEFT JOIN frame_spec_compatibility existing
  ON existing.frame_spec_id = fs.id
 AND existing.box_spec_id = bs.id
WHERE fs.active = 1
  AND fs.frame_type IN ('PARTITION', 'FEEDER')
  AND existing.id IS NULL
  AND (
    (fs.code LIKE '%_DEEP' AND bs.legacy_box_type = 'DEEP') OR
    (fs.code LIKE '%_SUPER' AND bs.legacy_box_type = 'SUPER') OR
    (fs.code LIKE '%_HORIZONTAL' AND bs.legacy_box_type = 'LARGE_HORIZONTAL_SECTION')
  );

-- +goose Down
DELETE c
FROM frame_spec_compatibility c
INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
WHERE fs.frame_type IN ('PARTITION', 'FEEDER');

UPDATE frame_specs
SET active = 0
WHERE frame_type IN ('PARTITION', 'FEEDER');
