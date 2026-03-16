-- +goose Up
ALTER TABLE apiaries
  ADD COLUMN type ENUM('STATIC', 'MOBILE') NOT NULL DEFAULT 'STATIC' AFTER name;

-- Ensure pre-existing rows are explicitly static.
UPDATE apiaries
SET type = 'STATIC'
WHERE type IS NULL OR type = '';

