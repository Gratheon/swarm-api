-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE hives
  ADD COLUMN collapse_date DATETIME NULL,
  ADD COLUMN collapse_cause TEXT NULL; 
