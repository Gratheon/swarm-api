-- +goose Up
ALTER TABLE `warehouse_modules`
MODIFY COLUMN `module_type` enum('DEEP','SUPER','ROOF','HORIZONTAL_FEEDER','QUEEN_EXCLUDER','BOTTOM') NOT NULL;

-- +goose Down
ALTER TABLE `warehouse_modules`
MODIFY COLUMN `module_type` enum('DEEP','SUPER','ROOF','QUEEN_EXCLUDER','BOTTOM') NOT NULL;
