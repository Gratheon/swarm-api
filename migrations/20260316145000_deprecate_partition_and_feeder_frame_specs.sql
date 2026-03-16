-- +goose Up
DELETE c
FROM frame_spec_compatibility c
INNER JOIN frame_specs fs ON fs.id = c.frame_spec_id
WHERE fs.frame_type IN ('PARTITION', 'FEEDER');

UPDATE frame_specs
SET active = 0
WHERE frame_type IN ('PARTITION', 'FEEDER');

-- +goose Down
UPDATE frame_specs
SET active = 1
WHERE frame_type IN ('PARTITION', 'FEEDER');
