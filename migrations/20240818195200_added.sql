-- +goose Up
UPDATE hives SET added = NOW() WHERE added IS NULL;