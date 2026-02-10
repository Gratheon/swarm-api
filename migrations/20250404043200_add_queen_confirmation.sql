-- +goose Up
ALTER TABLE frames_sides ADD COLUMN is_queen_confirmed BOOLEAN NOT NULL DEFAULT FALSE;
