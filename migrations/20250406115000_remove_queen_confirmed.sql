-- +goose Up
-- SQL in this section is executed when the migration is applied.
-- Remove the is_queen_confirmed column as this logic is moved to image-splitter
ALTER TABLE `frames_sides`
DROP COLUMN `is_queen_confirmed`;